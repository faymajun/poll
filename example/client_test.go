package main

import (
	"log"
	"net"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	client, err := net.Dial("tcp", ":2200")
	if err != nil {
		log.Println("err:", err.Error())
		return
	}

	client.Write([]byte("1234"))

	readBuf := make([]byte, 1024)
	for i := 0; i < 3; i++ {
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

// 异常退出
func TestClientAborted(t *testing.T) {
	client, err := net.Dial("tcp", ":2200")
	if err != nil {
		log.Println("dail err:", err.Error())
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
		_, err := net.Dial("tcp", ":2200")
		if err != nil {
			log.Println("dail err:", err.Error())
			return
		}
	}
}

func TestMoreClient(t *testing.T) {
	for i := 0; i < 100; i++ {
		go TestClient(nil)
	}
	time.Sleep(5 * time.Second)
}
