package main

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"

	"github.com/faymajun/poll"

	"github.com/pkg/errors"
)

func init() {
	poll.ConnOpened = func(c *poll.Conn) {
		log.Println(c.RemoteAddr().String(), " Conn open")

		go Process(c)
	}

	poll.ConnClosed = func(c *poll.Conn) {
		log.Println(c.RemoteAddr().String(), " Conn closed")
	}
}

func main() {
	poll.Serve(":2202")

	select {}
}

var reader io.Reader

func Process(c *poll.Conn) {
	ctx, _ := context.WithCancel(context.Background())
	readBuf := bytes.NewBuffer(make([]byte, 0, 4096))
	reader = bufio.NewReader(readBuf)
	for {
		select {
		case <-ctx.Done():
			return
		case in := <-c.ReadChan:
			readBuf.Write(in)
			// dataLen, err := c.readHead(reader)
			_, err := testReadHead(reader)
			if err != nil {
				return
			}

			c.Write([]byte("5678"))
		}
	}

}

func testReadHead(reader io.Reader) (dataLen uint32, err error) {
	lenData := make([]byte, 4)
	if _, err := io.ReadFull(reader, lenData); err != nil {
		return 0, errors.Errorf("read msg lenData error %v", err)
	}
	log.Println("testReadHead: ", string(lenData))
	return 0, nil
}
