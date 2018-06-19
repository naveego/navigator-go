package client

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/naveego/navigator-go/pipeline"
	"github.com/naveego/navigator-go/publishers/protocol"
	"github.com/naveego/navigator-go/publishers/server"
)

type DataPointCollector struct {
	addr         string
	clientServer publisherClientServer
}

func NewDataPointCollector(addr string) (DataPointCollector, error) {

	collector := DataPointCollector{
		addr: addr,
	}

	return collector, nil
}

// Start starts a goroutine which will accept datapoints over the collector's address.
// The collector will listen on the address provided to NewDataPointCollector.
// The JSON-RPC method prefix is "PublisherClient.".
func (d *DataPointCollector) Start(output chan<- []pipeline.DataPoint) error {

	listener, err := server.OpenListener(d.addr)
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			d.clientServer = publisherClientServer{
				output:   output,
				listener: listener,
			}

			server := rpc.NewServer()
			server.RegisterName("PublisherClient", &d.clientServer)

			codec := jsonrpc.NewServerCodec(conn)

			go server.ServeCodec(codec)
		}
	}()

	return nil
}

func (d *DataPointCollector) Stop() {
	var response protocol.DoneResponse
	_ = d.clientServer.Done(protocol.DoneRequest{}, &response)
}

type publisherClientServer struct {
	listener net.Listener
	output   chan<- []pipeline.DataPoint
}

// SendDataPoints accepts JSON-RPC calls from the publisher and passes them to the data collector's handler.
func (d *publisherClientServer) SendDataPoints(sendRequest protocol.SendDataPointsRequest, response *protocol.SendDataPointsResponse) error {

	*response = protocol.SendDataPointsResponse{}

	d.output <- sendRequest.DataPoints

	return nil
}

// Done stops the collector.
func (d *publisherClientServer) Done(doneRequest protocol.DoneRequest, response *protocol.DoneResponse) error {
	d.listener.Close()
	close(d.output)
	*response = protocol.DoneResponse{}
	return nil
}
