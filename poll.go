package poll

import (
	"context"
	"log"
	"net"
	"os"
	"sync/atomic"
	"syscall"
)

const maxConnNum = 10000

var s *server

type server struct {
	ln        *listener
	fdconns   map[int]*Conn
	connCount int32
	readBuf   []byte

	ctx    context.Context
	cancel context.CancelFunc
}

var pollObj = newPoll()

func Serve(addr string) error {
	var ln listener
	var err error

	if ln.ln, err = net.Listen("tcp", addr); err != nil {
		return err
	}

	if ln.f, err = ln.ln.(*net.TCPListener).File(); err != nil {
		ln.close()
		return err
	}

	ln.fd = int(ln.f.Fd())
	if err := syscall.SetNonblock(ln.fd, true); err != nil {
		return err
	}

	pollObj.addRead(ln.fd)
	s = &server{
		ln:      &ln,
		fdconns: make(map[int]*Conn),
		readBuf: make([]byte, 65535),
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	go pollObj.run(s.ctx)

	return nil
}

func (s *server) Close() {
	// close listener
	if s.ln != nil {
		s.ln.close()
	}

	// close loop and all connects
	s.cancel()

	for _, c := range s.fdconns {
		c.Close()
	}
	pollObj.close()
}

type listener struct {
	ln       net.Listener
	f        *os.File
	fd       int
	writeBuf []byte
}

func (ln *listener) close() {
	if ln.fd != 0 {
		syscall.Close(ln.fd)
	}
	if ln.f != nil {
		ln.f.Close()
	}
	if ln.ln != nil {
		ln.ln.Close()
	}
}

func Close() {
	s.Close()
}

func accept(fd int, p *poll) error {
	nfd, sa, err := syscall.Accept(fd)
	if err != nil {
		if err == syscall.EAGAIN {
			return nil
		}
		return err
	}

	if atomic.LoadInt32(&s.connCount) > maxConnNum {
		log.Println("Conn accept max")
		syscall.Close(nfd)
		return nil
	}

	if err := syscall.SetNonblock(nfd, true); err != nil {
		return err
	}
	c := newConn(nfd, sa, sockAddrToAddr(sa))
	atomic.AddInt32(&s.connCount, 1)

	s.fdconns[c.fd] = c
	p.addRead(c.fd)

	if ConnOpened != nil {
		ConnOpened(c)
	}
	return nil
}

func closeConn(c *Conn) error {
	delete(s.fdconns, c.fd)
	atomic.AddInt32(&s.connCount, -1)
	syscall.Close(c.fd)

	if ConnClosed != nil {
		ConnClosed(c)
	}
	return nil
}

func readConn(c *Conn) error {
	n, err := syscall.Read(c.fd, s.readBuf)
	if n == 0 || err != nil {
		if err == syscall.EAGAIN {
			return nil
		}
		c.Close()
		return nil
	}

	in := append([]byte{}, s.readBuf[:n]...)
	log.Println(c.remoteAddr.String(), " receive:", string(in))
	c.receive(in)
	return nil
}

func SendConn(c *Conn, data []byte) error {
	log.Println(c.remoteAddr.String(), " Send:", string(data))
	for len(data) > 0 {
		nn, err := syscall.Write(c.fd, data)
		if err != nil {
			return err
		}
		data = data[nn:]
	}
	return nil
}

func sockAddrToAddr(sa syscall.Sockaddr) net.Addr {
	var a net.Addr
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		a = &net.TCPAddr{
			IP:   append([]byte{}, sa.Addr[:]...),
			Port: sa.Port,
		}
	case *syscall.SockaddrInet6:
		var zone string
		if sa.ZoneId != 0 {
			if ifi, err := net.InterfaceByIndex(int(sa.ZoneId)); err == nil {
				zone = ifi.Name
			}
		}

		a = &net.TCPAddr{
			IP:   append([]byte{}, sa.Addr[:]...),
			Port: sa.Port,
			Zone: zone,
		}
	}
	return a
}
