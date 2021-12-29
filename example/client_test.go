package main

import (
	"log"
	"net"
	"testing"
	"time"

	"github.com/faymajun/poll/config"
)

func TestClient(t *testing.T) {
	client, err := net.Dial("tcp", ":2202")
	if err != nil {
		log.Println("err:", err.Error())
		return
	}

	client.Write([]byte("1234"))

	readBuf := make([]byte, 1024)
	for i := 0; i < 100; i++ {
		n, err := client.Read(readBuf)
		if n == 0 || err != nil {
			log.Println("client exit. err:", err.Error())
			return
		}
		log.Println("receive:", string(readBuf[0:n]))
		time.Sleep(time.Second)
		client.Write([]byte("1234"))
	}
	client.Close()
}

func TestWrite(t *testing.T) {
	client, err := net.Dial("tcp", ":2202")
	if err != nil {
		log.Println("err:", err.Error())
		return
	}

	recvLen := 0
	differentSize := 0
	go func() {
		readBuf := make([]byte, config.BufferSize)
		for {
			n, _ := client.Read(readBuf)
			log.Println("receive, len", n)
			for i := 1; i < n; i++ {
				if readBuf[i] != readBuf[i-1] {
					log.Println("different cur receiveSize:", recvLen+i, " different index:", i, "last len:", n-i)
					if differentSize > 0 {
						log.Println("different size:", differentSize)
						differentSize = 0
					} else {
						differentSize = n - i
					}
				}
			}
			if differentSize > 0 {
				differentSize += n
			}
			recvLen += n
		}
	}()

	for i := 0; i < 10000000; i++ {
		client.Write([]byte("1234"))
		client.Write([]byte("1234"))
		time.Sleep(1 * time.Second)
		log.Println("total receive size", recvLen)
		recvLen = 0
	}
	client.Close()
}

// 异常退出
func TestClientAborted(t *testing.T) {
	client, err := net.Dial("tcp", ":2202")
	if err != nil {
		// log.Println("dail err:", err.Error())
		return
	}

	client.Write([]byte("1234"))

	readBuf := make([]byte, 1024)
	n, err := client.Read(readBuf)
	if n == 0 || err != nil {
		log.Println("client exit. err:", err.Error())
		return
	}
	log.Println("receive:", string(readBuf[0:n]))
}

func TestMaxConn(t *testing.T) {
	for i := 0; i < 200; i++ {
		time.Sleep(time.Second)
		client, err := net.Dial("tcp", ":2202")
		if err != nil {
			log.Println("dail err:", err.Error())
			continue
		}
		client.Write([]byte("1234"))
		log.Println("send 1234")
	}
}

func TestMoreClient(t *testing.T) {
	for i := 0; i < 100; i++ {
		go TestClient(nil)
	}
	time.Sleep(5 * time.Second)
}
