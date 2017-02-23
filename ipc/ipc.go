package ipc

import (
	"github.com/naveego/navigator-go/jsonrpc"
)

type RequestHandler interface {
	HandleRequest(req jsonrpc.Request) (jsonrpc.Response, error)
}

type NotificationHandler interface {
	HandleNotification(req jsonrpc.Request) error
}
