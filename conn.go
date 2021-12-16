package poll

import (
	"net"
	"sync"
	"sync/atomic"
	"syscall"
)

var (
	ConnOpened func(c *Conn)
	ConnClosed func(c *Conn)
)

type Conn struct {
	fd         int
	sa         syscall.Sockaddr
	remoteAddr net.Addr
	closed     int32
	sendLock   sync.Mutex
	ReadChan   chan []byte
}

func newConn(fd int, sa syscall.Sockaddr, remoteAddr net.Addr) *Conn {
	c := &Conn{
		fd:         fd,
		sa:         sa,
		remoteAddr: remoteAddr,
		ReadChan:   make(chan []byte),
	}
	return c
}

func (c *Conn) receive(data []byte) {
	c.ReadChan <- data
}

func (c *Conn) Write(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	c.sendLock.Lock()
	defer c.sendLock.Unlock()
	return SendConn(c, data)
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *Conn) Close() {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		return
	}

	closeConn(c)
}
