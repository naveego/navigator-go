package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/naveego/api/types/pipeline"
	"github.com/naveego/navigator-go/publishers/protocol"
	"github.com/naveego/navigator-go/publishers/server"
)

var (
	interval = flag.Duration("interval", time.Second, "duration defining interval between publishes")
	times    = flag.Int("times", 10, "int defining number of publishes to do")
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

	srv := server.NewPublisherServer(addr, &publisherHandler{})

	err := srv.ListenAndServe()
	if err != nil {
		logrus.Fatal("Error shutting down server: ", err)
	}
}

type publisherHandler struct {
}

func (h *publisherHandler) TestConnection(request protocol.TestConnectionRequest) (protocol.TestConnectionResponse, error) {
	logrus.Debugf("TestConnection: %#v", request)

	return protocol.TestConnectionResponse{
		Success: true,
		Message: "OK",
	}, nil
}

func (h *publisherHandler) DiscoverShapes(request protocol.DiscoverShapesRequest) (protocol.DiscoverShapesResponse, error) {
	logrus.Debugf("DiscoverShapes: %#v", request)

	return protocol.DiscoverShapesResponse{
		Shapes: pipeline.ShapeDefinitions{
			pipeline.ShapeDefinition{
				Name:        "test-shape",
				Description: "test-shape description",
				Keys:        []string{"id"},
				Properties: []pipeline.PropertyDefinition{
					{
						Name: "id",
						Type: "number",
					},
					{
						Name: "name",
						Type: "string",
					},
				},
			},
		},
	}, nil
}

func (h *publisherHandler) Init(protocol.InitRequest) (protocol.InitResponse, error) {
	return protocol.InitResponse{
		Success: true,
		Message: "Initialized",
	}, nil
}

func (h *publisherHandler) Dispose(protocol.DisposeRequest) (protocol.DisposeResponse, error) {
	return protocol.DisposeResponse{
		Success: true,
		Message: "Disposed",
	}, nil
}

func (h *publisherHandler) Publish(request protocol.PublishRequest, toClient protocol.PublisherClient) (protocol.PublishResponse, error) {
	logrus.Debugf("Publish:\r\n  request: %#v\r\n  transport: %#v", request, toClient)

	go func() {
		for i := 0; i < *times; i++ {
			dp := pipeline.DataPoint{
				Repository: "vandelay",
				Entity:     "item",
				Source:     "test",
				Action:     pipeline.DataPointUpsert,
				KeyNames:   []string{"id"},
				Data: map[string]interface{}{
					"id":   i,
					"name": "John Doe",
				},
			}

			logrus.Debugf("Publishing (%s of %s): %#v", i, times, dp)

			toClient.SendDataPoints(protocol.SendDataPointsRequest{DataPoints: []pipeline.DataPoint{dp}})

			logrus.Debugf("Sleeping for %s", *interval)

			time.Sleep(*interval)
		}
	}()

	return protocol.PublishResponse{
		Success: true,
		Message: fmt.Sprintf("Expect %d items", *times),
	}, nil

}
