package poll

import (
	"net"
	"sync"
	"syscall"

	"github.com/pkg/errors"
)

var (
	ConnOpened func(c *Conn)
	ConnClosed func(c *Conn)
)

type Conn struct {
	fd         int
	sa         syscall.Sockaddr
	remoteAddr net.Addr
	sync.RWMutex
	isClosed bool
	ReadChan chan []byte
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
	c.RLock()
	defer c.RUnlock()
	if c.isClosed {
		return
	}

	c.ReadChan <- data
}

func (c *Conn) Write(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	c.RLock()
	defer c.RUnlock()
	if c.isClosed {
		return errors.New("conn is closed")
	}

	return SendConn(c, data)
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *Conn) Close() {
	c.Lock()
	defer c.Unlock()
	if c.isClosed {
		return
	}
	c.isClosed = true

	closeConn(c)
	close(c.ReadChan)
}
