package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/naveego/api/types/pipeline"
	"github.com/naveego/navigator-go/subscribers/protocol"
	"github.com/naveego/navigator-go/subscribers/server"
)

func main() {

	logrus.SetOutput(os.Stdout)

	if len(os.Args) < 2 {
		fmt.Println("Not enough arguments.")
		os.Exit(-1)
	}

	flag.Parse()

	addr := os.Args[1]

	logrus.SetLevel(logrus.DebugLevel)

	srv := server.NewSubscriberServer(addr, &subscriberHandler{})

	err := srv.ListenAndServe()
	if err != nil {
		logrus.Fatal("Error shutting down server: ", err)
	}
}

type subscriberHandler struct {
}

func (h *subscriberHandler) Init(request protocol.InitRequest) (protocol.InitResponse, error) {
	logrus.Debugf("Init: %#v", request)

	return protocol.InitResponse{
		Success: true,
		Message: "OK",
	}, nil
}

func (h *subscriberHandler) TestConnection(request protocol.TestConnectionRequest) (protocol.TestConnectionResponse, error) {
	logrus.Debugf("TestConnection: %#v", request)

	return protocol.TestConnectionResponse{
		Success: true,
		Message: "OK",
	}, nil
}

func (h *subscriberHandler) DiscoverShapes(request protocol.DiscoverShapesRequest) (protocol.DiscoverShapesResponse, error) {
	logrus.Debugf("DiscoverShapes: %#v", request)

	return protocol.DiscoverShapesResponse{
		Shapes: pipeline.ShapeDefinitions{
			pipeline.ShapeDefinition{
				Name:        "test-shape",
				Description: "test-shape description",
				Keys:        []string{"ID"},
				Properties: []pipeline.PropertyDefinition{
					{
						Name: "ID",
						Type: "number",
					},
					{
						Name: "Name",
						Type: "string",
					},
				},
			},
		},
	}, nil
}

func (h *subscriberHandler) ReceiveDataPoint(request protocol.ReceiveShapeRequest) (protocol.ReceiveShapeResponse, error) {
	logrus.Debugf("ReceiveDataPoint: %#v", request)

	return protocol.ReceiveShapeResponse{
		Success: true,
	}, nil
}

func (h *subscriberHandler) Dispose(request protocol.DisposeRequest) (protocol.DisposeResponse, error) {
	logrus.Debugf("Dispose: %#v", request)

	return protocol.DisposeResponse{
		Success: true,
	}, nil
}
