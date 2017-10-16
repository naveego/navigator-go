package client

import (
	"io"
	"net"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/naveego/navigator-go/subscribers/protocol"
	"github.com/naveego/navigator-go/subscribers/server"

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

type mockSubscriber struct {
	mock.Mock
}

func (m *mockSubscriber) TestConnection(request protocol.TestConnectionRequest) (protocol.TestConnectionResponse, error) {
	return protocol.TestConnectionResponse{
		Success: true,
		Message: "OK!",
	}, nil
}

func (m *mockSubscriber) Init(request protocol.InitRequest) (protocol.InitResponse, error) {
	return protocol.InitResponse{Success: true}, nil
}

func (m *mockSubscriber) ReceiveDataPoint(request protocol.ReceiveShapeRequest) (protocol.ReceiveShapeResponse, error) {
	ret := m.Called(request)
	resp := ret.Get(0).(protocol.ReceiveShapeResponse)
	return resp, ret.Error(1)
}

func (m *mockSubscriber) Dispose(request protocol.DisposeRequest) (protocol.DisposeResponse, error) {
	panic("not implemented")
}

func (m *mockSubscriber) DiscoverShapes(request protocol.DiscoverShapesRequest) (protocol.DiscoverShapesResponse, error) {
	panic("not implemented")
}

var (
	mockSubscriberInstance = &mockSubscriber{}
	addr                   = "tcp://127.0.0.1:54321"
)

func init() {

	go func() {
		srv := server.NewSubscriberServer(addr, mockSubscriberInstance)

		err := srv.ListenAndServe()
		if err != nil {
			logrus.Fatal("Error shutting down server: ", err)
		}
	}()

}

func Test_publisherProxy_TestConnection(t *testing.T) {

	Convey("should call method and get response", t, func() {
		var err error
		var sut protocol.Subscriber
		conn, err := net.Dial("tcp", "127.0.0.1:54321")
		sut, err = NewSubscriber(conn)

		result, err := sut.TestConnection(protocol.TestConnectionRequest{
			Settings: map[string]interface{}{},
		})

		Convey("should call method without error", func() {
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})

		defer conn.Close()
	})
}

func Test_subscriberProxy_ReceiveDataPoint(t *testing.T) {
	var conn io.ReadWriteCloser

	Convey("should call method and get response", t, func() {
		mockSubscriberInstance.Reset()

		expected := protocol.ReceiveShapeResponse{
			Success: true,
			Message: "Got request and responded!",
		}
		mockSubscriberInstance.When("ReceiveDataPoint", mock.Any).Return(expected, nil)

		var err error
		var sut protocol.Subscriber
		conn, err = net.Dial("tcp", "127.0.0.1:54321")
		sut, err = NewSubscriber(conn)

		actual, err := sut.ReceiveDataPoint(protocol.ReceiveShapeRequest{})
		ok, err := mockSubscriberInstance.Verify()

		Convey("should call method without error", func() {
			So(actual, ShouldResemble, expected)
			So(err, ShouldBeNil)
			So(ok, ShouldBeTrue)
		})

		defer conn.Close()
	})

}
