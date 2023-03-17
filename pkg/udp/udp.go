package udp

import (
	"context"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
)

type recorder struct {
	bufPool  sync.Pool
	port     uint
	filename string
}

func NewRecorder(port uint, filename string) *recorder {

	return &recorder{
		port:     port,
		filename: filename,
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
	f, err := os.Create(r.filename)
	if err != nil {
		return err
	}
	defer f.Close()

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
			f.Write(size)
			f.Write(msg)
			f.Write([]byte{'\n'})
		}

	}

	return errors.New("udp recording stopped")
}
