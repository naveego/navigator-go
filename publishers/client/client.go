package client

import (
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/naveego/navigator-go/publishers/protocol"
)

type publisherProxy struct {
	client      *rpc.Client
	replyToAddr string
}

type PublisherProxy interface {
	protocol.ShapeDiscoverer
	protocol.ConnectionTester
	Init(protocol.InitRequest) (protocol.InitResponse, error)
	Dispose(protocol.DisposeRequest) (protocol.DisposeResponse, error)
	Publish(protocol.PublishRequest) (protocol.PublishResponse, error)
}

// NewPublisher returns a protocol.Publisher proxy which
// communicates with a real publisher over the provided connection.
// The publisher must own the connection and must not be shared between goroutines.
func NewPublisher(conn io.ReadWriteCloser) (PublisherProxy, error) {

	publisherProxy := &publisherProxy{
		client: jsonrpc.NewClient(conn),
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

func (p *publisherProxy) Init(request protocol.InitRequest) (resp protocol.InitResponse, err error) {
	err = p.client.Call("Publisher.Init", request, &resp)
	return
}
func (p *publisherProxy) Dispose(request protocol.DisposeRequest) (resp protocol.DisposeResponse, err error) {
	err = p.client.Call("Publisher.Dispose", request, &resp)
	return
}

func (p *publisherProxy) Publish(request protocol.PublishRequest) (resp protocol.PublishResponse, err error) {
	err = p.client.Call("Publisher.Publish", request, &resp)
	return
}
