package server

import (
	"io"
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	"github.com/naveego/navigator-go/jsonrpc"
	pubproto "github.com/naveego/navigator-go/publishers/protocol"
)

type connection struct {
	srv  *PublisherServer
	conn net.Conn
}

func (c *connection) serve() {

	reqReader := jsonrpc.NewRequestReader(c.conn)
	resWriter := jsonrpc.NewResponseWriter(c.conn)

	for {
		req, hasHeaders, err := reqReader.ReadRequest()
		if err != nil {
			if err == io.EOF {
				break
			}

			logrus.Warn("could not read request: " + err.Error())
			continue
		}
		if !hasHeaders {
			logrus.Warn("No Headers")
			continue
		}

		isNotification := false

		var resp jsonrpc.Response

		switch req.Method {
		case "discoverShapes":
			resp = c.handleDiscoverShapes(req)
		case "testConnection":
			resp = c.handleTestConnection(req)
		}

		if err != nil {
			logrus.Warn("could not process method ["+req.Method+"]", err)
			continue
		}

		if !isNotification {
			resp.ID = req.ID
			err = resWriter.WriteResponse(resp)
			if err != nil {
				logrus.Warn("could not write response for method ["+req.Method+"]", err)
			}
		}
	}

}

func (c *connection) handleTestConnection(req jsonrpc.Request) jsonrpc.Response {
	h, i := c.srv.handler.(pubproto.ConnectionTester)
	if !i {
		return jsonrpc.Response{}
	}

	var testReq pubproto.TestConnectionRequest

	paramsRaw, ok := req.Params.(map[string]interface{})
	if !ok {
		return jsonrpc.NewParamsTypeErrprResponse()
	}

	err := mapstructure.Decode(paramsRaw, &testReq)
	if err != nil {
		logrus.Warn("params could not be decoded: ", err)
		return jsonrpc.NewDecodeParamsErrorResponse()
	}

	result, err := h.TestConnection(testReq)
	if err != nil {
		return jsonrpc.NewMethodInvocationError("error testing connection", nil)
	}

	return jsonrpc.Response{
		Result: result,
	}
}

func (c *connection) handleDiscoverShapes(req jsonrpc.Request) jsonrpc.Response {
	h, i := c.srv.handler.(pubproto.ShapeDiscoverer)
	if !i {
		return jsonrpc.Response{}
	}

	var discoverReq pubproto.DiscoverShapesRequest

	paramsRaw, ok := req.Params.(map[string]interface{})
	if !ok {
		return jsonrpc.NewInvalidParamsResponse("params was not a map[string]interface{}")
	}

	err := mapstructure.Decode(paramsRaw, &discoverReq)
	if err != nil {
		logrus.Warn("params could not be decoded: ", err)
		return jsonrpc.NewInvalidParamsResponse("params could not be decoded")
	}

	result, err := h.DiscoverShapes(discoverReq)
	if err != nil {
		return jsonrpc.NewErrorResponse(-32001, "method invocation error", nil)
	}

	return jsonrpc.Response{
		Result: result,
	}
}
