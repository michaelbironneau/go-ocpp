package soap

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"time"

	"github.com/michaelbironneau/go-ocpp/internal"
	"github.com/michaelbironneau/go-ocpp/internal/log"
	"github.com/michaelbironneau/go-ocpp/messages"

	"github.com/google/uuid"
)

type Client struct {
	url string
}

// NewClient creates new SOAP client instance
func NewClient(url string) *Client {
	return &Client{
		url: url,
	}
}

type CallOptionsFrom toSendFrom
type CallOptions struct {
	From              CallOptionsFrom
	ChargeBoxIdentity string
}

// Call performs HTTP POST request
func (s *Client) Call(soapAction string, request messages.Request, response messages.Response, options *CallOptions) error {
	if len(s.url) == 0 {
		return errors.New("no URL to request")
	}

	envelope := toSendEnvelope{XMLNS: "http://www.w3.org/2003/05/soap-envelope"}
	envelope.Header = &toSendHeader{
		XMLNS:     "http://www.w3.org/2005/08/addressing",
		Action:    "/" + soapAction,
		To:        s.url,
		MessageID: "uuid:" + uuid.New().String(),
	}
	if options != nil {
		envelope.Header.From = toSendFrom(options.From)
		envelope.Header.ChargeBoxIdentity = options.ChargeBoxIdentity
	}
	envelope.Body.Content = request
	buffer := new(bytes.Buffer)

	encoder := xml.NewEncoder(buffer)

	if err := encoder.Encode(envelope); err != nil {
		return err
	}

	if err := encoder.Flush(); err != nil {
		return err
	}
	// read bytes from xml encoder buffer
	rawReq := buffer.Bytes()

	// add <?xml version="1.0" encoding="UTF-8"?>
	rawReq = []byte(xml.Header + string(rawReq))

	log.Debug("Sending request with %d bytes", len(rawReq))
	log.Debug("Sending raw request: %s", string(rawReq))

	req, err := http.NewRequest("POST", s.url, buffer)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/soap+xml")
	req.Header.Add("Content-Type", "application/xml; charset=utf-8")

	req.Close = true

	tr := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, 8*time.Second)
		},
	}

	client := &http.Client{Transport: tr, Timeout: internal.DefaultRequestTimeout}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	rawbody, _ := ioutil.ReadAll(res.Body)

	if len(rawbody) == 0 && res.StatusCode != http.StatusOK {
		return errors.New("response returned with a non-OK status code: " + res.Status)
	}

	respEnvelope, err := Unmarshal(rawbody)
	if err != nil {
		return err
	}

	fault := respEnvelope.Body.Fault
	if fault != nil {
		return fault
	}

	respVal, contentVal := reflect.ValueOf(response), reflect.ValueOf(respEnvelope.Body.Content)
	if respVal.Type() != contentVal.Type() {
		panic("incompatible types:" + respVal.Type().String() + contentVal.Type().String())
	}

	respVal.Elem().Set(contentVal.Elem())

	return nil
}
