package protocol

import (
	"github.com/naveego/api/types/pipeline"
)

type InitializePublisherRequest struct {
	Publisher pipeline.PublisherInstance
}

type DiscoverShapesRequest struct {
	PublisherInstance pipeline.PublisherInstance `json:"instance" mapstructure:"instance"`
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
