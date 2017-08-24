package client

import (
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/naveego/navigator-go/subscribers/protocol"
)

type subscriberProxy struct {
	client *rpc.Client
	seq    int64
}

// NewSubscriber returns a protocol.Subscriber proxy which
// communicates with a real subscriber over the provided connection.
// The subscriber must own the connection and must not be shared between goroutines.
func NewSubscriber(conn io.ReadWriteCloser) (protocol.Subscriber, error) {

	//jsonClient := jsonrpc.NewClient(conn)

	subscriberProxy := &subscriberProxy{
		client: jsonrpc.NewClient(conn),
	}

	return subscriberProxy, nil
}

func (p *subscriberProxy) TestConnection(request protocol.TestConnectionRequest) (resp protocol.TestConnectionResponse, err error) {
	resp = protocol.TestConnectionResponse{}
	err = p.client.Call("Subscriber.TestConnection", request, &resp)
	return
}

func (p *subscriberProxy) Init(request protocol.InitRequest) (resp protocol.InitResponse, err error) {
	err = p.client.Call("Subscriber.Init", request, &resp)
	return
}

func (p *subscriberProxy) ReceiveDataPoint(request protocol.ReceiveShapeRequest) (resp protocol.ReceiveShapeResponse, err error) {
	err = p.client.Call("Subscriber.ReceiveDataPoint", request, &resp)
	return
}

func (p *subscriberProxy) Dispose(request protocol.DisposeRequest) (resp protocol.DisposeResponse, err error) {
	err = p.client.Call("Subscriber.Dispose", request, &resp)
	return
}

func (p *subscriberProxy) DiscoverShapes(request protocol.DiscoverShapesRequest) (resp protocol.DiscoverShapesResponse, err error) {
	err = p.client.Call("Subscriber.DiscoverShapes", request, &resp)
	return
}
