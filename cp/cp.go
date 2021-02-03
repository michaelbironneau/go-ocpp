package cp

import (
	"context"
	"fmt"
	"github.com/michaelbironneau/go-ocpp/cs"
	"net/http"

	"github.com/michaelbironneau/go-ocpp"
	"github.com/michaelbironneau/go-ocpp/internal/service"
	"github.com/michaelbironneau/go-ocpp/messages/v1x/csreq"
	"github.com/michaelbironneau/go-ocpp/messages/v1x/csresp"
	"github.com/michaelbironneau/go-ocpp/ws"
)

// CentralSystemMessageHandler handles the OCPP messages coming from the central system
type CentralSystemMessageHandler func(cprequest csreq.CentralSystemRequest) (csresp.CentralSystemResponse, error)

type ChargePoint interface {
	service.CentralSystem
	Identity() string

	// WS related
	Connection() *ws.Conn
	WaitConnect() <-chan struct{}
	WaitDisconnect() <-chan struct{}
}

type chargePoint struct {
	service.CentralSystem
	identity         string
	centralSystemURL string
	headers          http.Header
	version          ocpp.Version
	transport        ocpp.Transport
	conn             *ws.Conn
	ctx              context.Context
	connectedChan    chan struct{}
}

// Run the charge point on the given port
// and handles each incoming CentralSystemRequest
func New(ctx context.Context, identity, csURL string, version ocpp.Version, transport ocpp.Transport, port *string, headers http.Header, cshandler cs.ChargePointMessageHandler) (*chargePoint, error) {
	cp := &chargePoint{
		identity:         identity,
		centralSystemURL: csURL,
		version:          version,
		transport:        transport,
		ctx:              ctx,
		headers: headers,
		connectedChan:    make(chan struct{}),
	}
	if transport == ocpp.JSON {
		err := cp.getNewWebsocketConnection()
		if err != nil {
			return nil, fmt.Errorf("could not dial to central system: %w", err)
		}
		go cp.handleWebsocketConnection(cshandler)
	}
	if transport == ocpp.SOAP {
		// remove
	}
	return cp, nil
}

func (cp *chargePoint) Identity() string {
	return cp.identity
}
