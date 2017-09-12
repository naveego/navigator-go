// +build !windows

package cmd

import (
	"net"
	"strings"
	"time"
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

	return net.Listen(proto, addr)
}

var DefaultConnectionFactory ConnectionFactory = func(addr string) (net.Conn, error) {
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
type ConnectionFactory func(addr string) (net.Conn, error)

// ListenerFactory creates a connection from an address.
type ListenerFactory func(addr string) (net.Listener, error)
