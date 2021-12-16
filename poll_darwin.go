//go:build darwin
// +build darwin

package poll

import (
	"context"
	"log"
	"syscall"
)

type poll struct {
	fd      int
	changes []syscall.Kevent_t
}

func newPoll() *poll {
	fd, err := syscall.Kqueue()
	if err != nil {
		panic(err)
	}

	if _, err := syscall.Kevent(fd, []syscall.Kevent_t{{
		Ident:  0,
		Filter: syscall.EVFILT_USER,
		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
	}}, nil, nil); err != nil {
		panic(err)
	}

	return &poll{
		fd: fd,
	}
}

func (p *poll) addRead(fd int) {
	p.changes = append(p.changes,
		syscall.Kevent_t{Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_READ},
	)
}

func (p *poll) close() {
	syscall.Close(p.fd)
}

func (p *poll) run(ctx context.Context) {
	events := make([]syscall.Kevent_t, 128)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := syscall.Kevent(p.fd, p.changes, events, nil)
			if err != nil && err != syscall.EINTR {
				log.Println(err)
				return
			}

			p.changes = p.changes[:0]

			for i := 0; i < n; i++ {
				fd := int(events[i].Ident)
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
