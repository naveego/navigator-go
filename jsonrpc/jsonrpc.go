package jsonrpc

import "encoding/json"

type Request struct {
	ID     string
	Method string
	Params interface{}
}

type Response struct {
	ID     string
	Result interface{}
	Error  Error
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Pre-defined JSON RPC error codes
const ErrorParse = -32700
const ErrorInvalidRequest = -32600
const ErrorMethodNotFound = -32601
const ErrorInvalidParams = -32602
const ErrorInternalError = -32603

func NewErrorResponse(code int, message string, data interface{}) Response {
	return Response{
		Error: Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

func NewInvalidParamsResponse(msg string) Response {
	return NewErrorResponse(ErrorInvalidParams, msg, nil)
}

func NewParamsTypeErrprResponse() Response {
	return NewInvalidParamsResponse("params was not the correct type")
}

func NewDecodeParamsErrorResponse() Response {
	return NewInvalidParamsResponse("could not decode params")
}

func NewMethodInvocationError(msg string, data interface{}) Response {
	return NewErrorResponse(ErrorInternalError, msg, data)
}

func SerializeRequest(req Request) (string, error) {
	j := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  req.Method,
	}

	if req.Params != nil {
		j["params"] = req.Params
	}

	if req.ID != "" {
		j["id"] = req.ID
	}

	b, e := json.Marshal(j)
	return string(b), e
}

func SerializeResponse(res Response) (string, error) {
	j := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      res.ID,
	}

	if res.Error.Code != 0 {
		j["error"] = res.Error
	} else {
		j["result"] = res.Result
	}

	b, e := json.Marshal(j)
	return string(b), e
}
