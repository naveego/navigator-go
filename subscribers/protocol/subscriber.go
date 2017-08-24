package protocol

import (
	"github.com/naveego/api/types/pipeline"
)

type InitializeSubscriberRequest struct {
	Publisher pipeline.PublisherInstance
}

type DiscoverShapesRequest struct {
	SubscriberInstance pipeline.SubscriberInstance `json:"instance" mapstructure:"instance"`
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

type ReceiveShapeRequest struct {
	SubscriberInstance pipeline.SubscriberInstance `json:"instance" mapstructure:"instance"`
	Pipeline           pipeline.Pipeline           `json:"pipeline" mapstructure:"pipeline"`
	Shape              pipeline.ShapeDefinition    `json:"shape" mapstructure:"shape"`
	DataPoint          pipeline.DataPoint          `json:"data" mapstructure:"data"`
}

type ReceiveShapeResponse struct {
	Success bool   `json:"success" mapstructure:"success"`
	Message string `json:"message" mapstructure:"message"`
}

type DataPointReceiver interface {
	Init(request InitRequest) (InitResponse, error)
	ReceiveDataPoint(request ReceiveShapeRequest) (ReceiveShapeResponse, error)
	Dispose(request DisposeRequest) (DisposeResponse, error)
}

type Subscriber interface {
	ConnectionTester
	DataPointReceiver
	ShapeDiscoverer
}
