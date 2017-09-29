package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/naveego/api/types/pipeline"
	"github.com/naveego/navigator-go/subscribers/protocol"
	"github.com/naveego/navigator-go/subscribers/server"
	"github.com/sirupsen/logrus"
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

	logrus.WithField("listen-addr", addr).Info("Started console_subscriber")

	srv := server.NewSubscriberServer(addr, &subscriberHandler{})

	err := srv.ListenAndServe()
	if err != nil {
		logrus.Fatal("Error shutting down server: ", err)
	}
}

type subscriberHandler struct {
	prefix     string
	fileWriter io.WriteCloser
}

func (h *subscriberHandler) Init(request protocol.InitRequest) (protocol.InitResponse, error) {
	logrus.Debugf("Init: %#v", request)

	if request.Settings != nil {
		h.prefix = request.Settings["prefix"].(string)

		if fileName, ok := request.Settings["file"]; ok && fileName != "" {
			f, err := os.Create(fileName.(string))
			if err != nil {
				return protocol.InitResponse{
					Success: false,
					Message: "couldn't open file: " + err.Error(),
				}, err
			}
			h.fileWriter = f
		}

	}

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
	logrus.WithField("prefix", h.prefix).Debugf("ReceiveDataPoint: %#v", request)

	if h.fileWriter != nil {
		fmt.Fprintf(h.fileWriter, "%s - %#v", h.prefix, request.DataPoint)
		fmt.Fprintln(h.fileWriter)
	}

	return protocol.ReceiveShapeResponse{
		Success: true,
	}, nil
}

func (h *subscriberHandler) Dispose(request protocol.DisposeRequest) (protocol.DisposeResponse, error) {
	logrus.Debugf("Dispose: %#v", request)

	if h.fileWriter != nil {
		_ = h.fileWriter.Close()
	}

	return protocol.DisposeResponse{
		Success: true,
	}, nil
}
