package jsonrpc

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSerializeRequest(t *testing.T) {

	Convey("Given a Request", t, func() {
		str, _ := SerializeRequest(Request{
			Method: "callStuff",
		})
		msg := parseResult(str)

		Convey("should serialize the jsonrpc = '2.0' property", func() {
			So(msg["jsonrpc"], ShouldEqual, "2.0")
		})
		Convey("should serialize the method property", func() {
			So(msg["method"], ShouldEqual, "callStuff")
		})
		Convey("should serialize a params string properly", func() {
			str, _ := SerializeRequest(Request{
				Params: "hello",
			})
			msg := parseResult(str)
			So(msg["params"], ShouldEqual, "hello")
		})
		Convey("should serialize a params object properly", func() {
			str, _ := SerializeRequest(Request{
				Params: map[string]interface{}{
					"id": 1,
				},
			})
			msg := parseResult(str)
			p, _ := msg["params"].(map[string]interface{})
			So(p["id"], ShouldEqual, 1)
		})
		Convey("should serialize a params array properly", func() {
			str, _ := SerializeRequest(Request{
				Params: []interface{}{
					"string",
					"hi",
				},
			})
			msg := parseResult(str)
			p, _ := msg["params"].([]interface{})
			So(p[0], ShouldEqual, "string")
			So(p[1], ShouldEqual, "hi")
		})
		Convey("should not serialize the id property if empty string", func() {
			_, ok := msg["id"]
			So(ok, ShouldBeFalse)
		})
		Convey("should serialie the id property if not empty string", func() {
			str, _ := SerializeRequest(Request{
				ID: "12345",
			})
			msg := parseResult(str)
			So(msg["id"], ShouldEqual, "12345")
		})

	})
}

func TestSerializeResponse(t *testing.T) {

	Convey("Given a Success Response", t, func() {
		str, _ := SerializeResponse(Response{
			ID: "testid",
		})
		msg := parseResult(str)

		Convey("should serialize the jsonrpc = 2.0 property", func() {
			So(msg["jsonrpc"], ShouldEqual, "2.0")
		})
		Convey("should serialize the result property (string)", func() {
			str, _ := SerializeResponse(Response{
				Result: "test",
			})
			msg := parseResult(str)
			So(msg["result"], ShouldEqual, "test")
		})
		Convey("should not contain an error property", func() {
			_, ok := msg["error"]
			So(ok, ShouldBeFalse)
		})
		Convey("should serialize the id property", func() {
			So(msg["id"], ShouldEqual, "testid")
		})
	})

	Convey("Given an Error Response", t, func() {
		str, _ := SerializeResponse(Response{
			ID: "testid",
			Error: Error{
				Code:    ErrorParse,
				Message: "error message goes here",
			},
		})
		msg := parseResult(str)

		Convey("should serialize the jsonrpc = 2.0 property", func() {
			So(msg["jsonrpc"], ShouldEqual, "2.0")
		})
		Convey("should not contain a result property", func() {
			_, ok := msg["result"]
			So(ok, ShouldBeFalse)
		})
		Convey("should serialize the id property", func() {
			So(msg["id"], ShouldEqual, "testid")
		})
		Convey("should serialize the error property", func() {
			eRaw, _ := msg["error"]
			e, _ := eRaw.(map[string]interface{})
			So(e["code"], ShouldEqual, ErrorParse)
		})
	})
}

func parseResult(str string) map[string]interface{} {
	var j map[string]interface{}
	json.Unmarshal([]byte(str), &j)
	return j
}
