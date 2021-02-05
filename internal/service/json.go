package service

import (
	"github.com/michaelbironneau/go-ocpp/messages"
	"github.com/michaelbironneau/go-ocpp/ws"
)

type JSON struct {
	conn *ws.Conn
}

// NewJSON before calling it, you should
// be reading messages from the connection
// as to get the responses back:
// go func() {
// 	for {
// 		conn.ReadMessage()
// 	}
// }()
func NewJSON(conn *ws.Conn) *JSON {
	return &JSON{
		conn: conn,
	}
}

func (service *JSON) Send(chargerID string, req messages.Request) (messages.Response, error) {
	return service.conn.SendRequest(chargerID, req)
}
