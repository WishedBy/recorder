package udp

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
)

type NextReaderI interface {
	ReadNext() ([]byte, error)
}

type NextReader struct {
	Src io.Reader
}

func (u *NextReader) ReadNext() ([]byte, error) {
	sb := make([]byte, 4)
	n, err := u.Src.Read(sb)
	if err != nil {
		return []byte{}, err
	}
	size := binary.BigEndian.Uint32(sb)
	content := make([]byte, size)
	n, err = u.Src.Read(content)
	if err != nil {
		return []byte{}, err
	}
	if n != int(size) {
		return content, errors.New("Unable to read next item.")
	}
	nl := make([]byte, 1)
	n, err = u.Src.Read(nl)
	return content, err
}

type recorder struct {
	bufPool sync.Pool
	port    uint
	dst     io.Writer
}

func NewRecorder(port uint, dst io.Writer) *recorder {

	return &recorder{
		port: port,
		dst:  dst,
		bufPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 65536)
			},
		},
	}
}

func (r *recorder) ListenAndRecord(ctx context.Context) error {
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(int(r.port)))
	if err != nil {
		return err
	}

	connection, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	readCh := make(chan []byte)
	go func() {

		for run := true; run; {
			select {
			case <-ctx.Done():
				run = false
				continue
			default:
			}
			b := r.bufPool.Get().([]byte)
			n, _, err := connection.ReadFrom(b)
			if err != nil {
				log.Println(err)
			}
			msg := b[:n]
			r.bufPool.Put(b)
			readCh <- msg

		}
	}()

	for run := true; run; {
		select {
		case <-ctx.Done():
			run = false
			continue
		case msg := <-readCh:
			size := make([]byte, 4)
			binary.BigEndian.PutUint32(size, uint32(len(msg)))
			r.dst.Write(size)
			r.dst.Write(msg)
			r.dst.Write([]byte{'\n'})
		}

	}

	return errors.New("udp recording stopped")
}

func SendUDP(ctx context.Context, addr *net.UDPAddr, src NextReaderI) error {

	conn, _ := net.DialUDP("udp", nil, addr)

	var data []byte
	var err error
	for {
		data, err = src.ReadNext()
		if err != nil {
			break
		}
		select {
		case <-ctx.Done():
			return errors.New("stopped sending, context done")
		default:
		}
		conn.Write(data)
	}
	return err
}
