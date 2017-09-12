package client

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/naveego/api/types/pipeline"

	"github.com/Sirupsen/logrus"
	"github.com/maraino/go-mock"
	"github.com/naveego/navigator-go/publishers/protocol"
	"github.com/naveego/navigator-go/publishers/server"
	. "github.com/smartystreets/goconvey/convey"
)

type mockConn struct {
	mock.Mock
}

func (f *mockConn) Read(p []byte) (n int, err error) {
	ret := f.Called(p)
	return ret.Int(0), ret.Error(1)
}

func (f *mockConn) Write(p []byte) (n int, err error) {
	panic("not implemented")
}

func (f *mockConn) Close() error {
	panic("not implemented")
}

type mockHandler struct {
	mock.Mock
}

func (m *mockHandler) Init(request protocol.InitRequest) (protocol.InitResponse, error) {
	m.Called(request)
	return protocol.InitResponse{Success: true}, nil
}

func (m *mockHandler) Dispose(request protocol.DisposeRequest) (protocol.DisposeResponse, error) {
	m.Called(request)
	return protocol.DisposeResponse{Success: true}, nil
}

func (m *mockHandler) Publish(request protocol.PublishRequest, toClient protocol.PublisherClient) (protocol.PublishResponse, error) {
	m.Called(request, toClient)

	go func() {
		_, err := toClient.SendDataPoints(protocol.SendDataPointsRequest{
			DataPoints: []pipeline.DataPoint{
				{
					Repository: "vandelay",
					Entity:     "item",
					Source:     "test",
					Action:     pipeline.DataPointUpsert,
					KeyNames:   []string{"id"},
					Data: map[string]interface{}{
						"id":   1,
						"name": "John Doe",
						"dob":  "1990-01-01T12:00:00Z",
					},
				},
			},
		})

		if err != nil {
			panic(err)
		}

	}()

	return protocol.PublishResponse{
		Success: true,
	}, nil

}

func (m *mockHandler) TestConnection(request protocol.TestConnectionRequest) (protocol.TestConnectionResponse, error) {
	res := m.Called(request)
	return res.Get(0).(protocol.TestConnectionResponse), res.Error(1)
}

var (
	mockHandlerInstance = &mockHandler{}
	publisherAddr       = "tcp://127.0.0.1:51001"
	collectorAddr       = "tcp://127.0.0.1:51002"
	output              <-chan []pipeline.DataPoint
)

func init() {

	go func() {
		srv := server.NewPublisherServer(publisherAddr, mockHandlerInstance)

		err := srv.ListenAndServe()
		if err != nil {
			logrus.Fatal("Error shutting down server: ", err)
		}
	}()

	collector, err := NewDataPointCollector(collectorAddr)

	if err != nil {
		panic(err.Error())
	}

	output, err = collector.Start()

}

func Test_publisherProxy_TestConnection(t *testing.T) {

	Convey("should call method and get response", t, func() {
		mockHandlerInstance.Reset()
		expected := protocol.TestConnectionResponse{
			Success: true,
			Message: "Test Complete!",
		}
		mockHandlerInstance.When("TestConnection", mock.Any).Return(expected, nil)

		conn, err := net.Dial("tcp", strings.Split(publisherAddr, "://")[1])

		sut, err := NewPublisher(conn)

		actual, err := sut.TestConnection(protocol.TestConnectionRequest{
			Settings: map[string]interface{}{},
		})

		ok, err := mockHandlerInstance.Verify()

		Convey("should call method without error", func() {
			So(err, ShouldBeNil)
			So(ok, ShouldBeTrue)
			So(actual, ShouldResemble, expected)

		})

		defer conn.Close()
	})
}

func Test_publisherProxy_ReceiveDataPoint(t *testing.T) {

	Convey("should call method and publisher should begin pushing messages in response", t, func() {
		mockHandlerInstance.Reset()
		mockHandlerInstance.When("Publish", mock.Any, mock.Any).Times(1)

		var err error
		var sut PublisherProxy
		conn, err := net.Dial("tcp", strings.Split(publisherAddr, "://")[1])
		defer conn.Close()

		sut, err = NewPublisher(conn)

		Convey("proxy method should not have error", nil)
		resp, err := sut.Publish(protocol.PublishRequest{
			PublishToAddress: collectorAddr,
			ShapeName:        "test",
		})

		So(err, ShouldBeNil)
		So(resp.Success, ShouldBeTrue)

		Convey("method in publisher should have been called", nil)
		ok, err := mockHandlerInstance.Verify()
		So(err, ShouldBeNil)
		So(ok, ShouldBeTrue)

		Convey("collector should receive data point from publisher", nil)

		timeout := time.After(time.Millisecond * 100)
		select {
		case <-output:
			So(true, ShouldBeTrue)
		case <-timeout:
			So(false, ShouldBeTrue)
		}
	})
}

func TestNewSubscriber(t *testing.T) {

	Convey("Given a connection", t, func() {
		conn := mockConn{}
		Convey("should create a publisher", func() {
			got, err := NewPublisher(&conn)
			So(err, ShouldBeNil)
			So(got, ShouldNotBeNil)
		})
	})

}
