package server

import (
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"

	"github.com/Sirupsen/logrus"
)

type PublisherServer struct {
	Addr    string
	handler interface{}
}

func NewPublisherServer(addr string, handler interface{}) *PublisherServer {
	return &PublisherServer{
		Addr:    addr,
		handler: handler,
	}
}

func (srv *PublisherServer) ListenAndServe() error {
	addr := srv.Addr

	listener, err := OpenListener(addr)

	if err != nil {
		return err
	}

	return srv.Serve(listener)
}

func OpenListener(addr string) (net.Listener, error) {
	proto := "tcp"
	p := strings.Index(addr, "://")
	if p != -1 {
		proto = addr[:p]
		addr = addr[p+3:]
	}

	var l net.Listener
	var err error
	if proto == "namedpipes" {
		l, err = getNamedPipeListener(addr)
	} else {
		l, err = net.Listen(proto, addr)
	}

	return l, err
}

func (srv *PublisherServer) Serve(listener net.Listener) error {
	defer listener.Close()

	logrus.Infof("Listening for connections on %s", srv.Addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		logrus.Infof("Client connected")
		server := rpc.NewServer()
		wrapper := &wrapper{publisher: srv.handler}
		server.RegisterName("Publisher", wrapper)
		codec := jsonrpc.NewServerCodec(conn)
		server.ServeCodec(codec)
	}
}

func (s *PublisherServer) newConnection(conn net.Conn) *connection {
	return &connection{
		srv:  s,
		conn: conn,
	}
}

type ServerError struct {
	Code    int
	Message string
}

func (s *ServerError) Error() string {
	return fmt.Sprintf("[server] %d - %s", s.Code, s.Message)
}
