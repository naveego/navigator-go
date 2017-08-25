package server

import (
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"

	"github.com/Sirupsen/logrus"

	"github.com/naveego/api/types/pipeline"
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

func (w *wrapper) Publish(request protocol.PublishRequest, response *protocol.PublishResponse) (err error) {

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
		publishToURL, err := url.Parse(request.PublishToAddress)
		if err != nil {
			return fmt.Errorf("PublishToAddress '%s' was malformed: %s", err)
		}
		conn, err := net.Dial("tcp", publishToURL.Host)
		if err != nil {
			return err
		}

		client := jsonrpc.NewClient(conn)

		transport := &jsonrpcDataTransport{
			client: client,
		}

		// Now it's up to the publisher to go off and pump the datapoints.
		go s.Publish(request, transport)

		// We respond that we started the publisher.
		*response = protocol.PublishResponse{
			Success: true,
			Message: "Publisher started.",
		}
	}
	return nil
}

type jsonrpcDataTransport struct {
	client *rpc.Client
}

func (dt *jsonrpcDataTransport) Send(dataPoints []pipeline.DataPoint) error {

	err := dt.client.Call("DataPointCollector.ReceiveDataPoints", dataPoints, nil)

	return err
}
