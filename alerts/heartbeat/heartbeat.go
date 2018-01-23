package heartbeat

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/byuoitav/state-parsing/alerts/base"
	"github.com/byuoitav/state-parsing/alerts/device"
	"github.com/fatih/color"
)

const DeviceIndex = "oit-static-av-devices"

type HeartbeatAlertFactory struct {
}

func (h *HeartbeatAlertFactory) Run(loggingLevel int) (int, []base.Alert, error) {
	return Run(loggingLevel)
}

func Run(loggingLevel int) (int, []base.Alert, error) {

	if loggingLevel > 0 {
		log.Printf(color.HiGreenString("[Heartbeat-Lost] starting run"))
	}

	addr := fmt.Sprintf("%s/%s/_search", os.Getenv("ELK_ADDR"), DeviceIndex)

	resp, body, err := base.MakeELKRequest(addr, "POST", []byte(HeartbeatLostQuery), loggingLevel)
	if err != nil {
		//there's an error
		log.Printf(color.HiRedString("[Heartbeat-Lost] There was an error with the initial query: %v", err.Error()))
		return 0, []base.Alert{}, err
	}
	//take our response
	log.Printf("%v, %s", resp, body)

	hrresp := device.HeartbeatLostQueryResponse{}

	err = json.Unmarshal(body, &hrresp)
	if err != nil {
		log.Printf(color.HiRedString("Couldn't unmmarshal response: %v", err.Error()))
		return 0, []base.Alert{}, err
	}

	log.Printf("%#v", hrresp)

	return 0, []base.Alert{}, nil
}
