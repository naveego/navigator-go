package jsonrpc

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRequestBufferReadHeaders(t *testing.T) {

	Convey("Calling ReadHeaders with empty buffer", t, func() {
		buf := RequestBuffer{}
		headers, hasHeaders, err := buf.ReadHeaders()

		Convey("should return false for headers exist", func() {
			So(hasHeaders, ShouldBeFalse)
		})
		Convey("should return nil for error", func() {
			So(err, ShouldBeNil)
		})
		Convey("should return empty map", func() {
			So(headers, ShouldBeEmpty)
		})
	})

	Convey("Calling ReadHeaders with headers present", t, func() {
		buf := RequestBuffer{}
		buf.buffer = []byte("Content-Length: 300\r\nContent-Type: application/json \r\n\r\n This is some content")
		buf.index = len(buf.buffer)
		headers, hasHeaders, err := buf.ReadHeaders()

		Convey("should return true for headers exist", func() {
			So(hasHeaders, ShouldBeTrue)
		})
		Convey("should return nil for error", func() {
			So(err, ShouldBeNil)
		})
		Convey("should return the headers", func() {
			So(headers["Content-Length"], ShouldEqual, "300")
			So(headers["Content-Type"], ShouldEqual, "application/json")
		})
		Convey("should reset the buffer to include only content", func() {
			So(buf.buffer, ShouldResemble, []byte(" This is some content"))
		})
		Convey("should set the index to the length of the new buffer", func() {
			So(buf.index, ShouldEqual, 21)
		})
	})

	Convey("Calling ReadHeaders with invalid header", t, func() {
		buf := RequestBuffer{}
		buf.buffer = []byte("Content-Length 300\r\n\r\n")
		buf.index = len(buf.buffer)
		headers, hasHeaders, err := buf.ReadHeaders()

		Convey("should return an error", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("should return false for hasHeaders", func() {
			So(hasHeaders, ShouldBeFalse)
		})
		Convey("should return an empty headers value", func() {
			So(headers, ShouldBeEmpty)
		})
	})

	Convey("Calling Append", t, func() {
		Convey("should add the data to the buffer", func() {
			buf := RequestBuffer{}
			So(len(buf.buffer), ShouldEqual, 0)
			buf.Append([]byte("0123456789"))
			So(len(buf.buffer), ShouldEqual, 10)
		})
		Convey("should set the index to the length of the buffer", func() {
			buf := RequestBuffer{}
			So(buf.index, ShouldEqual, 0)
			buf.Append([]byte("0123456789"))
			So(buf.index, ShouldEqual, 10)
		})
	})

}

func TestRequestBufferReadContent(t *testing.T) {

	Convey("Calling ReadContent", t, func() {
		content := []byte(" 0123456789")
		contentLength := len(content)
		buf := RequestBuffer{}
		buf.buffer = content
		buf.index = len(buf.buffer)

		Convey("should return the content as a string", func() {
			c, _ := buf.ReadContent(contentLength)
			So(string(c), ShouldEqual, string(" 0123456789"))
		})
	})
}
