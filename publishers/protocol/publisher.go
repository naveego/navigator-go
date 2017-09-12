package protocol

import (
	"github.com/naveego/api/types/pipeline"
)

type DiscoverShapesRequest struct {
	Settings map[string]interface{} `json:"settings"`
}

type DiscoverShapesResponse struct {
	Shapes pipeline.ShapeDefinitions `json:"shapes"`
}

type ShapeDiscoverer interface {
	DiscoverShapes(request DiscoverShapesRequest) (DiscoverShapesResponse, error)
}

type TestConnectionRequest struct {
	Settings map[string]interface{} `json:"settings"`
}

type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ConnectionTester interface {
	TestConnection(request TestConnectionRequest) (TestConnectionResponse, error)
}

type PublishRequest struct {
	ShapeName        string `json:"shapeName"`
	PublishToAddress string `json:"publishToAddress" mapstructure:"publishToAddress"`
}

type PublishResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DataPublisher interface {
	Init(InitRequest) (InitResponse, error)
	Dispose(DisposeRequest) (DisposeResponse, error)
	Publish(request PublishRequest, toClient PublisherClient) (PublishResponse, error)
}

type PublishDataNotification struct {
	DataPoints []pipeline.DataPoint `json:"data_points" mapstructure:"data_points"`
}

type InitRequest struct {
	Settings map[string]interface{} `json:"settings"`
}

type InitResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DisposeRequest struct{}
type DisposeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// PublisherClient is the interface the publisher sends data points to.
type PublisherClient interface {
	// SendDataPoints sends data points to the client.
	SendDataPoints(sendRequest SendDataPointsRequest) (SendDataPointsResponse, error)
	// Done tells the client that the publisher is done sending data points for now.
	Done(DoneRequest) (DoneResponse, error)
}

type SendDataPointsRequest struct {
	DataPoints []pipeline.DataPoint
}

type SendDataPointsResponse struct {
}
type DoneRequest struct{}

type DoneResponse struct{}
