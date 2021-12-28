package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/faymajun/poll"
)

var (
	port   int
	socket string
)

func init() {
	flag.IntVar(&port, "p", 2202, "listen port")
	flag.StringVar(&socket, "s", "/tmp/sendfd.sock", "socket")
}

func main() {
	flag.Parse()

	if !flag.Parsed() || socket == "" {
		flag.Usage()
		os.Exit(1)
	}

	tcpl, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}
	defer tcpl.Close()

	var c net.Conn
	c, err = tcpl.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	var f *os.File
	f, err = c.(*net.TCPConn).File()
	if err != nil {
		log.Fatal(err)
	}

	os.Remove(socket)
	var l net.Listener
	l, err = net.Listen("unix", socket)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	var a net.Conn
	a, err = l.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	listenConn := a.(*net.UnixConn)
	if err = poll.Put(listenConn, f); err != nil {
		log.Fatal(err)
	}
}
