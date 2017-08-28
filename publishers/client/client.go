package client

import (
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/naveego/api/types/pipeline"
	"github.com/naveego/navigator-go/publishers/protocol"
)

type publisherProxy struct {
	client      *rpc.Client
	replyToAddr string
}

type PublisherProxy interface {
	protocol.ShapeDiscoverer
	protocol.ConnectionTester
	Publish(instance pipeline.PublisherInstance, shape pipeline.ShapeDefinition) error
}

// NewPublisher returns a protocol.Publisher proxy which
// communicates with a real publisher over the provided connection.
// The publisher must own the connection and must not be shared between goroutines.
func NewPublisher(conn io.ReadWriteCloser, replyToAddr string) (PublisherProxy, error) {

	publisherProxy := &publisherProxy{
		client:      jsonrpc.NewClient(conn),
		replyToAddr: replyToAddr,
	}

	return publisherProxy, nil
}

func (p *publisherProxy) DiscoverShapes(request protocol.DiscoverShapesRequest) (resp protocol.DiscoverShapesResponse, err error) {
	err = p.client.Call("Publisher.DiscoverShapes", request, &resp)
	return
}

func (p *publisherProxy) TestConnection(request protocol.TestConnectionRequest) (resp protocol.TestConnectionResponse, err error) {
	err = p.client.Call("Publisher.TestConnection", request, &resp)
	return
}

func (p *publisherProxy) Publish(instance pipeline.PublisherInstance, shape pipeline.ShapeDefinition) error {
	dummy := protocol.PublishResponse{}

	request := protocol.PublishRequest{
		PublishedShape:    shape,
		PublisherInstance: instance,
		PublishToAddress:  p.replyToAddr,
	}

	err := p.client.Call("Publisher.Publish", request, &dummy)

	return err
}
