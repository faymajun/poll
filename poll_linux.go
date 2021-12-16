package poll

import (
	"context"
	"log"
	"syscall"
)

type poll struct {
	fd     int
	wakeFd int
}

func newPoll() *poll {
	l := new(poll)
	fd, err := syscall.EpollCreate1(0)
	if err != nil {
		panic(err)
	}

	l.fd = fd
	r1, _, err0 := syscall.Syscall(syscall.SYS_EVENTFD2, 0, 0, 0)
	if err0 != 0 {
		syscall.Close(l.fd)
		panic("syscall.SYS_EVENTFD2 err")
	}

	l.wakeFd = int(r1)
	l.addRead(l.wakeFd)
	return l
}

func (p *poll) addRead(fd int) {
	err := syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_ADD, fd,
		&syscall.EpollEvent{
			Fd:     int32(fd),
			Events: syscall.EPOLLIN,
		})
	if err != nil {
		panic(err)
	}
}

func (p *poll) close() {
	syscall.Close(p.wakeFd)
	syscall.Close(p.fd)
}

func (p *poll) run(ctx context.Context) {
	events := make([]syscall.EpollEvent, 128)
	wakeBuf := make([]byte, 8)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := syscall.EpollWait(p.fd, events, 100)
			if err != nil && err != syscall.EINTR {
				log.Println(err)
				return
			}

			for i := 0; i < n; i++ {
				fd := int(events[i].Fd)
				if fd == p.wakeFd {
					syscall.Read(p.wakeFd, wakeBuf)
					continue
				}
				c := s.fdconns[fd]
				switch {
				case c == nil:
					if err := accept(fd, p); err != nil {
						log.Println("accept fd:", fd, " error:", err.Error())
						return
					}
				default:
					if err := readConn(c); err != nil {
						log.Println("readConn Conn:", c, "error:", err.Error())
						return
					}
				}
			}
		}
	}
}
