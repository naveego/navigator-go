package server

import (
	"net"

	"github.com/Microsoft/go-winio"
)

func getNamedPipeListener(addr string) (net.Listener, error) {
	return winio.ListenPipe(addr, nil)
}
