package poll

import (
	"log"
	"net"
	"os"
	"syscall"

	"github.com/faymajun/poll/config"
)

var playerBin []byte

func init() {
	playerBin = make([]byte, config.PlayerSize)
	for i := 0; i < config.PlayerSize; i++ {
		playerBin[i] = 'a'
	}
}

func Get(via *net.UnixConn, num int, filenames []string) ([]*os.File, error) {
	if num < 1 {
		return nil, nil
	}

	// get the underlying socket
	viaf, err := via.File()
	if err != nil {
		return nil, err
	}
	socket := int(viaf.Fd())
	defer viaf.Close()

	// recvmsg
	buf := make([]byte, syscall.CmsgSpace(num*4))
	log.Println("len buf", len(buf))
	p := make([]byte, config.BufferSize)
	n, oobn, _, _, err := syscall.Recvmsg(socket, p, buf, 0)
	log.Println("syscall.Recvmsg", n, oobn, p[n-1], buf, err)
	n2, oobn2, _, _, err := syscall.Recvmsg(socket, p, buf, 0)
	log.Println("syscall.Recvmsg 2 ", n2, oobn2, p[n2-1], buf, err)
	addr, _ := net.ResolveTCPAddr("tcp", string(buf))
	log.Println("addr", addr.String())
	if err != nil {
		return nil, err
	}

	// parse control msgs
	var msgs []syscall.SocketControlMessage
	msgs, err = syscall.ParseSocketControlMessage(buf)

	// convert fds to files
	res := make([]*os.File, 0, len(msgs))
	for i := 0; i < len(msgs) && err == nil; i++ {
		var fds []int
		fds, err = syscall.ParseUnixRights(&msgs[i])
		log.Println("get fds", fds)
		for fi, fd := range fds {
			log.Println("get fi", fi, "fd", fd)
			var filename string
			if fi < len(filenames) {
				filename = filenames[fi]
			}

			res = append(res, os.NewFile(uintptr(fd), filename))
		}
	}

	return res, err
}

func Put(via *net.UnixConn, files ...*os.File) error {
	if len(files) == 0 {
		return nil
	}

	viaf, err := via.File()
	if err != nil {
		return err
	}
	socket := int(viaf.Fd())
	defer viaf.Close()

	fds := make([]int, len(files))
	for i := range files {
		fds[i] = int(files[i].Fd())
		log.Println("put fd", fds[i])
	}

	log.Println("put fds", fds)
	rights := syscall.UnixRights(fds...)
	log.Println("syscall.sendmgs buf", rights)
	return syscall.Sendmsg(socket, playerBin, rights, nil, 0)
}
