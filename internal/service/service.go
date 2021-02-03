package service

import "github.com/michaelbironneau/go-ocpp/messages"

type Service interface {
	Send(request messages.Request) (messages.Response, error)
}
