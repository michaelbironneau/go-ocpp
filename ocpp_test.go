package ocpp_test

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/eduhenke/go-ocpp"
	"github.com/eduhenke/go-ocpp/cp"
	"github.com/eduhenke/go-ocpp/cs"
	"github.com/eduhenke/go-ocpp/messages/v1x/cpreq"
	"github.com/eduhenke/go-ocpp/messages/v1x/cpresp"
	"github.com/eduhenke/go-ocpp/messages/v1x/csreq"
	"github.com/eduhenke/go-ocpp/messages/v1x/csresp"
)

func Test_Connection(t *testing.T) {
	cpointConnected := make(chan string)
	cpointDisconnected := make(chan string)

	csysPort := ":5050"
	csys := cs.New()
	csys.SetChargePointConnectionListener(func(cpID string) {
		t.Log("cpoint connected: ", cpID)
		cpointConnected <- cpID
	})
	csys.SetChargePointDisconnectionListener(func(cpID string) {
		t.Log("cpoint disconnected: ", cpID)
		cpointDisconnected <- cpID
	})
	go csys.Run(csysPort, func(req cpreq.ChargePointRequest, cpID string) (cpresp.ChargePointResponse, error) {
		return nil, errors.New("not supported")
	})

	csysURL := "ws://localhost" + csysPort
	t.Run("one chargepoint", func(t *testing.T) {
		cpID := "123"
		t.Run("single time", func(t *testing.T) {
			disconnect := testConnectionDisconnection(t, cpID, csysURL, cpointConnected, cpointDisconnected)
			disconnect()
		})

		t.Run("multiple times", func(t *testing.T) {
			attempts := 100
			for i := 0; i <= attempts; i++ {
				disconnect := testConnectionDisconnection(t, cpID, csysURL, cpointConnected, cpointDisconnected)
				disconnect()
			}
		})
	})

	t.Run("multiple chargepoints", func(t *testing.T) {
		t.Run("shuffled connect + disconnect", func(t *testing.T) {
			a := "chargerA"
			b := "chargerB"
			c := "chargerC"

			disconnectA := testConnectionDisconnection(t, a, csysURL, cpointConnected, cpointDisconnected)
			disconnectB := testConnectionDisconnection(t, b, csysURL, cpointConnected, cpointDisconnected)
			disconnectC := testConnectionDisconnection(t, c, csysURL, cpointConnected, cpointDisconnected)

			disconnectB()
			disconnectA()
			disconnectC()
		})

		t.Run("large chargers pool, shuffled connect + disconnect", func(t *testing.T) {
			chargerPoolSize := 100
			chargerPool := make([]struct {
				id         string
				disconnect func()
			}, chargerPoolSize)

			// creating and connecting chargers
			for i := 0; i < chargerPoolSize; i++ {
				cpID := strconv.Itoa(i)
				disconnect := testConnectionDisconnection(t, cpID, csysURL, cpointConnected, cpointDisconnected)
				chargerPool[i] = struct {
					id         string
					disconnect func()
				}{
					id:         cpID,
					disconnect: disconnect,
				}
			}

			rand.Shuffle(chargerPoolSize, func(i, j int) {
				chargerPool[i], chargerPool[j] = chargerPool[j], chargerPool[i]
			})
			for i, charger := range chargerPool {
				t.Log(i, charger.id)
				charger.disconnect()
			}
		})
	})
}

func testConnectionDisconnection(t *testing.T, cpID, csysURL string, cpointConnected, cpointDisconnected chan string) (disconnect func()) {
	cpctx, killCp := context.WithCancel(context.Background())
	cpoint, err := cp.New(cpctx, cpID, csysURL, ocpp.V16, ocpp.JSON)
	if err != nil {
		t.Fatal(fmt.Errorf("chargepoint could not start: %w", err))
	}

	go cpoint.Run(nil, func(req csreq.CentralSystemRequest) (csresp.CentralSystemResponse, error) {
		return nil, errors.New("not supported")
	})

	connectedCpID := <-cpointConnected
	if cpoint.Identity() != connectedCpID {
		t.Log("correct charge point did not connect")
		cpointConnected <- connectedCpID
	}

	done := make(chan struct{})
	go func() {
		for disconnectedCpID := range cpointDisconnected {
			if cpoint.Identity() != disconnectedCpID {
				// t.Logf("correct charge point did not disconnect. charge point expected (%s), charge point disconnected(%s)", cpoint.Identity(), disconnectedCpID)
				cpointDisconnected <- disconnectedCpID
			} else {
				t.Log("correctly disconnected charge point: " + cpoint.Identity())
				done <- struct{}{}
				return
			}
		}
	}()

	return func() {
		killCp()
		<-done
	}
}
