// +build !windows

package server

import (
	"errors"
	"io"
	"net"
	"strings"
	"time"
)

func getNamedPipeListener(addr string) (net.Listener, error) {
	return nil, errors.New("Named pipes is only supported on Windows")
}

var DefaultConnectionFactory ConnectionFactory = func(addr string) (io.ReadWriteCloser, error) {
	timeout := time.Second * 5
	proto := "tcp"
	p := strings.Index(addr, "://")
	if p != -1 {
		proto = addr[:p]
		addr = addr[p+3:]
	}

	return net.DialTimeout(proto, addr, timeout)
}

// ConnectionFactory creates a connection from an address.
type ConnectionFactory func(addr string) (io.ReadWriteCloser, error)
