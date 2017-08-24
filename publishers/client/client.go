package client

import (
	"fmt"
	"io"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/naveego/api/types/pipeline"
	"github.com/naveego/navigator-go/publishers/protocol"
)

type publisherProxy struct {
	client    *rpc.Client
	replyAddr string
}

type PublisherProxy interface {
	protocol.ShapeDiscoverer
	protocol.ConnectionTester
	Publish(instance pipeline.PublisherInstance, shape pipeline.ShapeDefinition) (chan []pipeline.DataPoint, error)
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

func (p *publisherProxy) Publish(instance pipeline.PublisherInstance, shape pipeline.ShapeDefinition) (chan []pipeline.DataPoint, error) {
	dummy := protocol.PublishResponse{}

	// TODO: Make this work correctly. For some reason net.Listen is blocking and then something panics.
	var err error
	var conn net.Conn
	var port string
	for i := 51000; i < 51100; i++ {
		port = fmt.Sprintf("127.0.0.1:%v", i)
		fmt.Printf("testing port %s", port)
		fmt.Println()
		listener, err := net.Listen("tcp", port)
		if err == nil {
			fmt.Printf("using port %s", port)
			fmt.Println()
			conn, err = listener.Accept()
			if err != nil {
				fmt.Printf("accepting on port %s", port)
				fmt.Println()

				break
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("couldn't get a port for datapoint listener, %v", err)
	}

	handler := &datapointHandler{
		output: make(chan []pipeline.DataPoint, 100),
	}

	server := rpc.NewServer()
	server.RegisterName("DataPointHandler", handler)

	codec := jsonrpc.NewServerCodec(conn)

	go server.ServeCodec(codec)

	request := protocol.PublishRequest{
		PublishedShape:    shape,
		PublisherInstance: instance,
		PublishToAddress:  fmt.Sprintf("tcp://localhost:%v", 1),
	}

	err = p.client.Call("Publisher.Publish", request, &dummy)

	return nil, err
}

type datapointHandler struct {
	output chan []pipeline.DataPoint
}

func (d *datapointHandler) ReceiveDataPoints(datapoints []pipeline.DataPoint, ok *bool) error {
	d.output <- datapoints
	*ok = true
	return nil
}
