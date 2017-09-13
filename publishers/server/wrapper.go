package server

import (
	"io"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
	"time"

	winio "github.com/Microsoft/go-winio"
	"github.com/Sirupsen/logrus"

	"github.com/naveego/navigator-go/publishers/protocol"
)

// wrapper adapts the protocol.* interfaces to the pattern required by net/rpc/jsonrpc.
type wrapper struct {
	publisher interface{}
}

func (w *wrapper) DiscoverShapes(request protocol.DiscoverShapesRequest, response *protocol.DiscoverShapesResponse) (err error) {
	if s, ok := w.publisher.(protocol.ShapeDiscoverer); ok {
		r, err := s.DiscoverShapes(request)
		*response = r
		return err
	}
	return nil
}

func (w *wrapper) TestConnection(request protocol.TestConnectionRequest, response *protocol.TestConnectionResponse) (err error) {
	if s, ok := w.publisher.(protocol.ConnectionTester); ok {
		r, err := s.TestConnection(request)
		*response = r
		return err
	}
	return nil
}

func (w *wrapper) Init(request protocol.InitRequest, response *protocol.InitResponse) (err error) {
	if s, ok := w.publisher.(protocol.DataPublisher); ok {
		r, err := s.Init(request)
		*response = r
		return err
	}
	return nil
}
func (w *wrapper) Dispose(request protocol.DisposeRequest, response *protocol.DisposeResponse) (err error) {
	if s, ok := w.publisher.(protocol.DataPublisher); ok {
		r, err := s.Dispose(request)
		*response = r
		return err
	}
	return nil
}

func (w *wrapper) Publish(request protocol.PublishRequest, response *protocol.PublishResponse) (err error) {
	logrus.Info("Calling Publish")
	*response = protocol.PublishResponse{
		Success: false,
	}

	if s, ok := w.publisher.(protocol.DataPublisher); ok {

		// Here we create a JSON-RPC client that
		// will take the datapoints produced by the
		// publisher implementation and transport them
		// back to the publication manager for dispatch
		// to the pipeline.

		logrus.Debugf("PublishToAddress was %s", request.PublishToAddress)
		conn, err := DefaultConnectionFactory(request.PublishToAddress)
		if err != nil {
			return err
		}

		client := jsonrpc.NewClient(conn)

		transport := &jsonrpcDataTransport{
			client: client,
		}

		// Now it's up to the publisher to go off and pump the datapoints.
		go func() {
			defer conn.Close()
			s.Publish(request, transport)
		}()

		// We respond that we started the publisher.
		*response = protocol.PublishResponse{
			Success: true,
			Message: "Publisher started.",
		}
	} else {

		// We respond that we started the publisher.
		*response = protocol.PublishResponse{
			Success: false,
			Message: "Handler doesn't implement DataPublisher.",
		}
	}

	return nil
}

type jsonrpcDataTransport struct {
	client *rpc.Client
}

func (dt *jsonrpcDataTransport) SendDataPoints(request protocol.SendDataPointsRequest) (resp protocol.SendDataPointsResponse, err error) {

	err = dt.client.Call("PublisherClient.SendDataPoints", request, &resp)

	return
}

func (dt *jsonrpcDataTransport) Done(request protocol.DoneRequest) (resp protocol.DoneResponse, err error) {

	err = dt.client.Call("PublisherClient.Done", request, &resp)

	return
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
