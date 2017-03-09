package server

import (
	"fmt"
	"net"
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
		rw, err := listener.Accept()
		if err != nil {
			return err
		}

		logrus.Infof("Client connected")
		conn := srv.newConnection(rw)
		go conn.serve()
	}
}

func (s *SubscriberServer) newConnection(conn net.Conn) *connection {
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
