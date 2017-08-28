package client

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/naveego/api/types/pipeline"
	"github.com/naveego/navigator-go/publishers/server"
)

type DataPointCollector struct {
	addr     string
	listener net.Listener
	handler  func([]pipeline.DataPoint) error
}

func NewDataPointCollector(addr string, handler func([]pipeline.DataPoint) error) (DataPointCollector, error) {

	collector := DataPointCollector{
		addr:    addr,
		handler: handler,
	}

	return collector, nil
}

// Start starts a goroutine which will accept datapoints over the collector's address.
func (d *DataPointCollector) Start() error {

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

			server := rpc.NewServer()
			server.RegisterName("DataPointCollector", d)

			codec := jsonrpc.NewServerCodec(conn)

			go server.ServeCodec(codec)
		}
	}()

	return nil
}

// ReceiveDataPoints accepts JSON-RPC calls from the publisher and passes them to the data collector's handler.
func (d *DataPointCollector) ReceiveDataPoints(datapoints []pipeline.DataPoint, ok *bool) error {
	err := d.handler(datapoints)
	*ok = true
	if err != nil {
		d.Stop()
		return err
	}
	return nil
}

// Stop stops the collector.
func (d *DataPointCollector) Stop() {
	d.listener.Close()
}
