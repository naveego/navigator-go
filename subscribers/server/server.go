package server

import (
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"

	"github.com/Sirupsen/logrus"
)

type SubscriberServer struct {
	Addr    string
	handler interface{}
}

func NewSubscriberServer(addr string, handler interface{}) *SubscriberServer {
	return &SubscriberServer{
		Addr:    addr,
		handler: handler,
	}
}

func (srv *SubscriberServer) ListenAndServe() error {
	proto := "tcp"
	addr := srv.Addr

	p := strings.Index(srv.Addr, "://")
	if p != -1 {
		proto = srv.Addr[:p]
		addr = srv.Addr[p+3:]
	}

	var l net.Listener
	var err error
	if proto == "namedpipes" {
		l, err = getNamedPipeListener(addr)
	} else {
		l, err = net.Listen(proto, addr)
	}

	if err != nil {
		return err
	}

	return srv.Serve(l)
}

func (srv *SubscriberServer) Serve(listener net.Listener) error {
	defer listener.Close()

	logrus.Infof("Listening for connections on %s", srv.Addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		logrus.Infof("Client connected")
		server := rpc.NewServer()
		wrapper := &wrapper{subscriber: srv.handler}
		server.RegisterName("Subscriber", wrapper)
		codec := jsonrpc.NewServerCodec(conn)
		server.ServeCodec(codec)
	}
}

type ServerError struct {
	Code    int
	Message string
}

func (s *ServerError) Error() string {
	return fmt.Sprintf("[server] %d - %s", s.Code, s.Message)
}
