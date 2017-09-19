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

func main() {

	logrus.SetOutput(os.Stdout)

	if len(os.Args) < 1 {
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
	inited   bool
	count    int
	interval time.Duration
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

func (h *publisherHandler) Init(request protocol.InitRequest) (protocol.InitResponse, error) {
	var err error

	h.count, _ = request.Settings["count"].(int)
	intervalString := request.Settings["interval"].(string)
	if intervalString != "" {
		h.interval, err = time.ParseDuration(intervalString)
		if err != nil {
			return protocol.InitResponse{
				Success: false,
				Message: "Invalid interval in settings.",
			}, err
		}
	}

	if h.count > 0 && h.interval > 0 {
		h.inited = true

		return protocol.InitResponse{
			Success: true,
			Message: fmt.Sprintf("Initialized. Will send %d items with an interval of %.2f seconds.", h.count, h.interval.Seconds()),
		}, nil

	}
	return protocol.InitResponse{
		Success: false,
		Message: "Invalid count or interval in settings.",
	}, fmt.Errorf("invalid count or interval in settings: count=%v, interval=%v", h.count, h.interval)
}

func (h *publisherHandler) Dispose(protocol.DisposeRequest) (protocol.DisposeResponse, error) {

	h.inited = false
	h.count = 0
	h.interval = time.Duration(0)

	return protocol.DisposeResponse{
		Success: true,
		Message: "Disposed",
	}, nil
}

func (h *publisherHandler) Publish(request protocol.PublishRequest, toClient protocol.PublisherClient) (protocol.PublishResponse, error) {
	logrus.Debugf("Publish:\r\n  request: %#v\r\n  transport: %#v", request, toClient)

	go func() {
		for i := 0; i < h.count; i++ {
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

			logrus.Debugf("Publishing (%s of %s): %#v", i, h.count, dp)

			toClient.SendDataPoints(protocol.SendDataPointsRequest{DataPoints: []pipeline.DataPoint{dp}})

			logrus.Debugf("Sleeping for %.2f seconds", h.interval.Seconds())

			time.Sleep(h.interval)
		}
	}()

	return protocol.PublishResponse{
		Success: true,
		Message: fmt.Sprintf("Expect %d items", h.count),
	}, nil

}
