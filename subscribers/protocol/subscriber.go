package protocol

import (
	"github.com/naveego/api/types/pipeline"
)

type InitializeSubscriberRequest struct {
	Settings map[string]interface{} `json:"settings"`
}

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

type InitRequest struct {
	Settings map[string]interface{}  `json:"settings"`
	Mappings []pipeline.ShapeMapping `json:"mappings"`
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
	ShapeName string             `json:"shape_name" mapstructure:"shape"`
	DataPoint pipeline.DataPoint `json:"data" mapstructure:"data"`
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
