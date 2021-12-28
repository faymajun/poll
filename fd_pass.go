package poll

import (
	"net"
	"os"
	"syscall"
)

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
	_, _, _, _, err = syscall.Recvmsg(socket, nil, buf, 0)
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

		for fi, fd := range fds {
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
	}

	rights := syscall.UnixRights(fds...)
	return syscall.Sendmsg(socket, nil, rights, nil, 0)
}
