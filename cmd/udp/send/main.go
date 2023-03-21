package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/WishedBy/recorder/pkg/udp"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		fmt.Println(<-c)
		cancel()

	}()

	ip := flag.String("ip", "127.0.0.1", "ip to send to")
	port := flag.Uint("port", 514, "udp port to send to")
	fileP := flag.String("file", "last", "file to send. could be last, first, [suffix] or file path")
	flag.Parse()
	file := *fileP
	spew.Dump(file)
	files, err := filepath.Glob("./data/udp-*.log")
	if err != nil {
		log.Fatal(err)
	}
	sort.Slice(files, func(i, j int) bool {
		return strings.Compare(files[i], files[j]) == -1
	})
	switch file {
	case "last":
		file = files[len(files)-1]
	case "first":
		file = files[len(files)-1]
	default:
		if _, err := os.Stat("./data/udp-" + file + ".log"); err == nil {
			file = "./data/udp-" + file + ".log"
		}
	}
	f, err := os.Open(file)

	udpAddr, err := net.ResolveUDPAddr("udp", *ip+":"+strconv.Itoa(int(*port)))
	if err != nil {
		log.Fatal(err)
	}
	err = udp.SendUDP(ctx, udpAddr, &udp.NextReader{Src: f})
	log.Println(err)

}
