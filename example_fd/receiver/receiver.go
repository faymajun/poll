package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/faymajun/poll/config"

	"github.com/faymajun/poll"
)

var (
	socket string
	listb  []byte
)

func init() {
	listb = make([]byte, config.BufferSize)
	for i := 0; i < config.BufferSize; i++ {
		listb[i] = '2'
	}
	flag.StringVar(&socket, "s", "/tmp/sendfd.sock", "socket")
}

func main() {
	flag.Parse()

	if !flag.Parsed() || socket == "" {
		flag.Usage()
		os.Exit(1)
	}

	c, err := net.Dial("unix", socket)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	fdConn := c.(*net.UnixConn)

	time.Sleep(time.Second * 3)

	var fs []*os.File
	fs, err = poll.Get(fdConn, 1, []string{"a file"})
	if err != nil {
		log.Fatal(err)
	}
	f := fs[0]
	defer f.Close()

	for {
		b := make([]byte, 4096)
		var n int
		n, err = f.Read(b)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		log.Printf("%s", b[:n])
		f.Write(listb)
	}
}
