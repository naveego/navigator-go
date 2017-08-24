package client

import (
	"net"
	"testing"
	"time"

	"github.com/naveego/api/types/pipeline"

	"github.com/Sirupsen/logrus"
	pubapi "github.com/naveego/api/pipeline/publisher"
	"github.com/naveego/navigator-go/publishers/protocol"
	"github.com/naveego/navigator-go/publishers/server"

	"github.com/maraino/go-mock"
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
}

func (m *mockHandler) TestConnection(request protocol.TestConnectionRequest) (protocol.TestConnectionResponse, error) {
	res := m.Called(request)
	return res.Get(0).(protocol.TestConnectionResponse), res.Error(1)
}

var (
	mockHandlerInstance = &mockHandler{}
	addr                = "tcp://127.0.0.1:51000"
)

func init() {

	go func() {
		srv := server.NewPublisherServer(addr, mockHandlerInstance)

		err := srv.ListenAndServe()
		if err != nil {
			logrus.Fatal("Error shutting down server: ", err)
		}
	}()
}

func Test_publisherProxy_TestConnection(t *testing.T) {

	Convey("should call method and get response", t, func() {
		mockHandlerInstance.Reset()
		mockHandlerInstance.When("TestConnection", mock.Any).Return(protocol.TestConnectionResponse{
			Success: true,
		}, nil)

		conn, err := net.Dial("tcp", "127.0.0.1:51000")

		sut, err := NewPublisher(conn)

		_, err = sut.TestConnection(protocol.TestConnectionRequest{
			Settings: map[string]interface{}{},
		})

		ok, err := mockHandlerInstance.Verify()

		Convey("should call method without error", func() {
			So(err, ShouldBeNil)
			So(ok, ShouldBeTrue)

		})

		defer conn.Close()
	})
}

func Test_publisherProxy_ReceiveDataPoint(t *testing.T) {

	Convey("should call method and publisher should begin pushing messages in response", t, func() {
		mockHandlerInstance.Reset()
		mockHandlerInstance.When("Publish", mock.Any, mock.Any)

		var err error
		var sut PublisherProxy
		conn, err := net.Dial("tcp", "127.0.0.1:11111")

		sut, err = NewPublisher(conn)

		output, err := sut.Publish(pipeline.PublisherInstance{
			Description: "d",
			Name:        "n",
			Settings:    map[string]interface{}{},
			Shapes:      pipeline.ShapeDefinitions{},
		}, pipeline.ShapeDefinition{})

		ok, err := mockHandlerInstance.Verify()

		Convey("should call method without error", func() {
			So(err, ShouldBeNil)
			So(ok, ShouldBeTrue)

			Convey("should see datapoint from publisher", func() {

				timeout := time.After(time.Second * 5)

				select {
				case <-output:
					So(true, ShouldBeTrue)
				case <-timeout:
					So(false, ShouldBeTrue)
				}
			})
		})

		defer conn.Close()
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
