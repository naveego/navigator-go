package server

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/Sirupsen/logrus"

	"github.com/naveego/api/types/pipeline"
	"github.com/naveego/navigator-go/publishers/protocol"
)

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

func (w *wrapper) Publish(request protocol.PublishRequest, response *protocol.PublishResponse) (err error) {

	*response = protocol.PublishResponse{
		Success: false,
	}

	if s, ok := w.publisher.(protocol.DataPublisher); ok {

		logrus.Debugf("PublishToAddress was %s", request.PublishToAddress)
		conn, err := net.Dial("tcp", request.PublishToAddress)
		if err != nil {
			return err
		}

		client := jsonrpc.NewClient(conn)

		transport := &jsonrpcDataTransport{
			client: client,
		}

		s.Publish(request, transport)
		*response = protocol.PublishResponse{
			Success: true,
		}
	}
	return nil
}

type jsonrpcDataTransport struct {
	client *rpc.Client
}

func (dt *jsonrpcDataTransport) Send(dataPoints []pipeline.DataPoint) error {

	err := dt.client.Call("DataPointHandler.ReceiveDataPoints", dataPoints, nil)

	return err
}
