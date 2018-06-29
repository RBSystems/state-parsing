package heartbeat

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/byuoitav/state-parsing/alerts"
	"github.com/byuoitav/state-parsing/alerts/base"
	"github.com/byuoitav/state-parsing/alerts/device"
)

/*
const DeviceIndex = "oit-static-av-devices"

type LostHeartbeatAlertFactory struct {
	alerts.AlertFactory
}

func (h *LostHeartbeatAlertFactory) Run() error {
	log.L.Infof("Starting run")

	addr := fmt.Sprintf("%s/%s/_search", os.Getenv("ELK_ADDR"), DeviceIndex)

	respCode, body, err := base.MakeELKRequest(addr, "POST", []byte(HeartbeatLostQuery), h.LogLevel)
	if err != nil {
		h.Error("error with the initial query: %s", err)
		return err
	}
	if respCode/100 != 2 {
		msg := fmt.Sprintf("[lost-heartbeat] Non 200 response received from the initial query: %v, %s", respCode, body)
		log.L.Error(msg)
		return errors.New(msg)
	}
	hrresp := device.HeartbeatLostQueryResponse{}

	err = json.Unmarshal(body, &hrresp)
	if err != nil {
		log.L.Error("couldn't unmarshal response: %s", err)
		return err
	}

	//process the alerts
	h.AlertsToSend, err = processHeartbeatLostResponse(hrresp)
	return err
}
*/

type RestoredHeartbeatAlertFactory struct {
	alerts.AlertFactory
}

func (h *RestoredHeartbeatAlertFactory) Init() {
	h.Name = names.HEARTBEAT_RESTORED
	h.LogLevel = logger.VERBOSE
}

func (h *RestoredHeartbeatAlertFactory) Run() error {
	h.Info("Starting run")

	addr := fmt.Sprintf("%s/%s/_search", os.Getenv("ELK_ADDR"), DeviceIndex)

	respCode, body, err := base.MakeELKRequest(addr, "POST", []byte(HeartbeatRestoredQuery), h.LogLevel)
	if err != nil {
		h.Error("error with initial query: %s", err)
		return err
	}

	if respCode/100 != 2 {
		msg := fmt.Sprintf("non 200 response received from the initial query: %v, %s", respCode, body)
		h.Error(msg)
		return errors.New(msg)
	}

	hrresp := device.HeartbeatRestoredQueryResponse{}

	err = json.Unmarshal(body, &hrresp)
	if err != nil {
		h.Error("couldn't unmarshal response: %s", err)
		return err
	}

	// process the alerts
	h.AlertsToSend, err = processHeartbeatRestoredResponse(hrresp)
	return err
}
