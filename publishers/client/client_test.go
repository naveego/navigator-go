package client

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/naveego/api/types/pipeline"

	"github.com/Sirupsen/logrus"
	"github.com/maraino/go-mock"
	pubapi "github.com/naveego/api/pipeline/publisher"
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

func (m *mockHandler) Publish(request protocol.PublishRequest, transport pubapi.DataTransport) {
	m.Called(request, transport)

	transport.Send([]pipeline.DataPoint{
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
	})
}

func (m *mockHandler) TestConnection(request protocol.TestConnectionRequest) (protocol.TestConnectionResponse, error) {
	res := m.Called(request)
	return res.Get(0).(protocol.TestConnectionResponse), res.Error(1)
}

var (
	mockHandlerInstance = &mockHandler{}
	publisherAddr       = "tcp://127.0.0.1:51001"
	collectorAddr       = "tcp://127.0.0.1:51002"
	output              = make(chan []pipeline.DataPoint, 100)
)

func init() {

	go func() {
		srv := server.NewPublisherServer(publisherAddr, mockHandlerInstance)

		err := srv.ListenAndServe()
		if err != nil {
			logrus.Fatal("Error shutting down server: ", err)
		}
	}()

	collector, err := NewDataPointCollector(collectorAddr, func(d []pipeline.DataPoint) error {
		output <- d
		return nil
	})

	if err != nil {
		panic(err.Error())
	}

	collector.Start()

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

		sut, err := NewPublisher(conn, collectorAddr)

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

		sut, err = NewPublisher(conn, collectorAddr)

		Convey("proxy method should not have error", nil)
		err = sut.Publish(pipeline.PublisherInstance{
			Description: "d",
			Name:        "n",
			Settings:    map[string]interface{}{},
			Shapes:      pipeline.ShapeDefinitions{},
		}, pipeline.ShapeDefinition{})

		So(err, ShouldBeNil)

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

		defer conn.Close()

	})
}

func TestNewSubscriber(t *testing.T) {

	Convey("Given a connection", t, func() {
		conn := mockConn{}
		Convey("should create a publisher", func() {
			got, err := NewPublisher(&conn, collectorAddr)
			So(err, ShouldBeNil)
			So(got, ShouldNotBeNil)
		})
	})

}
