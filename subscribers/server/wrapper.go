package server

import (
	"github.com/naveego/navigator-go/subscribers/protocol"
)

// wrapper adapts the protocol.* interfaces to the pattern required by net/rpc/jsonrpc.
type wrapper struct {
	subscriber interface{}
}

func (w *wrapper) TestConnection(request protocol.TestConnectionRequest, response *protocol.TestConnectionResponse) error {
	if s, ok := w.subscriber.(protocol.ConnectionTester); ok {
		r, err := s.TestConnection(request)
		*response = r
		return err
	}
	return nil
}

func (w *wrapper) Init(request protocol.InitRequest, response *protocol.InitResponse) (err error) {
	if s, ok := w.subscriber.(protocol.DataPointReceiver); ok {
		r, err := s.Init(request)
		*response = r
		return err
	}
	return nil
}

func (w *wrapper) ReceiveDataPoint(request protocol.ReceiveShapeRequest, response *protocol.ReceiveShapeResponse) (err error) {
	if s, ok := w.subscriber.(protocol.DataPointReceiver); ok {
		r, err := s.ReceiveDataPoint(request)
		*response = r
		return err
	}
	return nil
}

func (w *wrapper) Dispose(request protocol.DisposeRequest, response *protocol.DisposeResponse) (err error) {
	if s, ok := w.subscriber.(protocol.DataPointReceiver); ok {
		r, err := s.Dispose(request)
		*response = r
		return err
	}
	return nil
}

func (w *wrapper) DiscoverShapes(request protocol.DiscoverShapesRequest, response *protocol.DiscoverShapesResponse) (err error) {
	if s, ok := w.subscriber.(protocol.ShapeDiscoverer); ok {
		r, err := s.DiscoverShapes(request)
		*response = r
		return err
	}
	return nil
}
