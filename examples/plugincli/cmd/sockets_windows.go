package cmd

import (
	"io"
	"net"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"
)

// DefaultListenerFactory opens a connection to addr (which should have the form "scheme://address/to/medium")
// Supported schemes are tcp and unix on Linux/OSX, tcp and namedpipes on Windows.
var DefaultListenerFactory ListenerFactory = func(addr string) (net.Listener, error) {
	proto := "tcp"
	p := strings.Index(addr, "://")
	if p != -1 {
		proto = addr[:p]
		addr = addr[p+3:]
	}

	var l net.Listener
	var err error
	if proto == "namedpipes" {
		l, err = winio.ListenPipe(addr, nil)
	} else {
		l, err = net.Listen(proto, addr)
	}

	return l, err
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

// ListenerFactory creates a connection from an address.
type ListenerFactory func(addr string) (net.Listener, error)
