package server

import (
	"fmt"
	"net"

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
	l, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	return srv.Serve(l)
}

func (srv *PublisherServer) Serve(listener net.Listener) error {
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
