// +build !windows

package server

import (
	"errors"
	"net"
)

func getNamedPipeListener(addr string) (net.Listener, error) {
	return nil, errors.New("Named pipes is only supported on Windows")
}
