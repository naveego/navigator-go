package server

import (
	"io"
	"net"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"
)

func getNamedPipeListener(addr string) (net.Listener, error) {
	return winio.ListenPipe(addr, nil)
}

var DefaultConnectionFactory ConnectionFactory = func(addr string) (io.ReadWriteCloser, error) {
	timeout := time.Second * 5
	proto := "tcp"
	p := strings.Index(addr, "://")
	if p != -1 {
		proto = addr[:p]
		addr = addr[p+3:]
	}

	if proto == "namedpipes" {
		return winio.DialPipe(addr, &timeout)
	}

	return net.DialTimeout(proto, addr, timeout)

}

// ConnectionFactory creates a connection from an address.
type ConnectionFactory func(addr string) (io.ReadWriteCloser, error)
