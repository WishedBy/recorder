package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/WishedBy/recorder/pkg/udp"
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

	port := flag.Uint("port", 514, "udp port to listen on")
	flag.Parse()
	file := "./data/udp-" + time.Now().UTC().Format("20060102T150405") + ".log"

	f, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	rec := udp.NewRecorder(*port, f)
	err = rec.ListenAndRecord(ctx)
	log.Println(err)

}
