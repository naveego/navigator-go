package server

import (
	"io"
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	"github.com/naveego/navigator-go/jsonrpc"
	subproto "github.com/naveego/navigator-go/subscribers/protocol"
)

type connection struct {
	srv  *SubscriberServer
	conn net.Conn
}

func (c *connection) serve() {

	reqReader := jsonrpc.NewRequestReader(c.conn)
	resWriter := jsonrpc.NewResponseWriter(c.conn)

	for {
		req, hasHeaders, err := reqReader.ReadRequest()
		if err != nil {
			if err == io.EOF {
				logrus.Info("Client disconnected")
				break
			}

			logrus.Warnf("could not read request: (%T) %v", err, err)
			continue
		}
		if !hasHeaders {
			logrus.Warn("No Headers")
			continue
		}

		isNotification := false

		var resp jsonrpc.Response

		switch req.Method {
		case "init":
			resp = c.handleInit(req)
		case "discoverShapes":
			resp = c.handleDiscoverShapes(req)
		case "testConnection":
			resp = c.handleTestConnection(req)
		case "receiveDataPoint":
			resp = c.handleReceiveShape(req)
		case "dispose":
			resp = c.handleDispose(req)
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

func (c *connection) handleInit(req jsonrpc.Request) jsonrpc.Response {
	h, i := c.srv.handler.(subproto.DataPointReceiver)
	if !i {
		return jsonrpc.Response{}
	}

	paramsRaw, ok := req.Params.(map[string]interface{})
	if !ok {
		return jsonrpc.NewParamsTypeErrprResponse()
	}

	var initReq subproto.InitRequest

	err := mapstructure.Decode(paramsRaw, &initReq)
	if err != nil {
		logrus.Warn("params could not be decoded: ", err)
		return jsonrpc.NewDecodeParamsErrorResponse()
	}

	result, err := h.Init(initReq)
	if err != nil {
		logrus.Warn("error executing method: ", err)
		return jsonrpc.NewMethodInvocationError("error initializing subscriber ", err.Error())
	}

	return jsonrpc.Response{
		Result: result,
	}
}

func (c *connection) handleTestConnection(req jsonrpc.Request) jsonrpc.Response {
	h, i := c.srv.handler.(subproto.ConnectionTester)
	if !i {
		return jsonrpc.Response{}
	}

	var testReq subproto.TestConnectionRequest

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
		logrus.Warn("error executing method: ", err)
		return jsonrpc.NewMethodInvocationError("error testing connection ", err.Error())
	}

	return jsonrpc.Response{
		Result: result,
	}
}

func (c *connection) handleDiscoverShapes(req jsonrpc.Request) jsonrpc.Response {
	h, i := c.srv.handler.(subproto.ShapeDiscoverer)
	if !i {
		return jsonrpc.Response{}
	}

	var discoverReq subproto.DiscoverShapesRequest

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
		return jsonrpc.NewErrorResponse(-32001, "method invocation error", err.Error())
	}

	return jsonrpc.Response{
		Result: result,
	}
}

func (c *connection) handleReceiveShape(req jsonrpc.Request) jsonrpc.Response {
	h, i := c.srv.handler.(subproto.DataPointReceiver)
	if !i {
		return jsonrpc.Response{}
	}

	var receiveReq subproto.ReceiveShapeRequest

	paramsRaw, ok := req.Params.(map[string]interface{})
	if !ok {
		return jsonrpc.NewInvalidParamsResponse("params was not a map[string]interface{}")
	}

	err := mapstructure.Decode(paramsRaw, &receiveReq)
	if err != nil {
		logrus.Warn("params could not be decoded: ", err)
		return jsonrpc.NewInvalidParamsResponse("params could not be decoded")
	}

	result, err := h.ReceiveDataPoint(receiveReq)
	if err != nil {
		return jsonrpc.NewErrorResponse(-32001, "method invocation error", err.Error())
	}

	return jsonrpc.Response{
		Result: result,
	}
}

func (c *connection) handleDispose(req jsonrpc.Request) jsonrpc.Response {
	h, i := c.srv.handler.(subproto.DataPointReceiver)
	if !i {
		return jsonrpc.Response{}
	}

	paramsRaw, ok := req.Params.(map[string]interface{})
	if !ok {
		return jsonrpc.NewParamsTypeErrprResponse()
	}

	var dispReq subproto.DisposeRequest

	err := mapstructure.Decode(paramsRaw, &dispReq)
	if err != nil {
		logrus.Warn("params could not be decoded: ", err)
		return jsonrpc.NewDecodeParamsErrorResponse()
	}

	result, err := h.Dispose(dispReq)
	if err != nil {
		logrus.Warn("error executing method: ", err)
		return jsonrpc.NewMethodInvocationError("error initializing subscriber ", err.Error())
	}

	return jsonrpc.Response{
		Result: result,
	}
}
