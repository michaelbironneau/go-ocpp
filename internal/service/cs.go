package service

import (
	"github.com/michaelbironneau/go-ocpp/messages/v1x/cpreq"
	"github.com/michaelbironneau/go-ocpp/messages/v1x/cpresp"
	"github.com/michaelbironneau/go-ocpp/messages/v1x/csreq"
	"github.com/michaelbironneau/go-ocpp/messages/v1x/csresp"
	"github.com/michaelbironneau/go-ocpp/ws"
)

type CentralSystem interface {
	Send(request csreq.CentralSystemRequest) (csresp.CentralSystemResponse, error)
}

type CentralSystemSOAP struct {
	*SOAP
}


func (service *CentralSystemSOAP) Send(req cpreq.ChargePointRequest) (cpresp.ChargePointResponse, error) {
	rawResp, err := service.SOAP.Send(req)
	if err != nil {
		return nil, err
	}
	resp, ok := rawResp.(cpresp.ChargePointResponse)
	if !ok {
		return nil, cpresp.ErrorNotChargePointResponse
	}
	return resp, nil
}

type CentralSystemJSON struct {
	*JSON
}

func NewCentralSystemJSON(conn *ws.Conn) CentralSystem {
	return &CentralSystemJSON{NewJSON(conn)}
}

func (service *CentralSystemJSON) Send(request csreq.CentralSystemRequest) (csresp.CentralSystemResponse, error) {
	rawResp, err := service.JSON.Send(request)
	if err != nil {
		return nil, err
	}
	resp, ok := rawResp.(csresp.CentralSystemResponse)
	if !ok {
		return nil, cpresp.ErrorNotChargePointResponse
	}
	return resp, nil
}
