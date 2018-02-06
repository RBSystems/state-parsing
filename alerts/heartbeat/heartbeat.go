package heartbeat

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/byuoitav/state-parsing/alerts/base"
	"github.com/byuoitav/state-parsing/alerts/device"
	"github.com/fatih/color"
)

const DeviceIndex = "oit-static-av-devices"

type LostHeartbeatAlertFactory struct {
}

func (h *LostHeartbeatAlertFactory) Run(loggingLevel int) (map[string][]base.Alert, error) {
	return Run(loggingLevel)
}

func Run(loggingLevel int) (map[string][]base.Alert, error) {

	if loggingLevel > 0 {
		log.Printf(color.HiGreenString("[lost-heartbeat] starting run"))
	}

	addr := fmt.Sprintf("%s/%s/_search", os.Getenv("ELK_ADDR"), DeviceIndex)

	respCode, body, err := base.MakeELKRequest(addr, "POST", []byte(HeartbeatLostQuery), loggingLevel)
	if err != nil {
		//there's an error
		log.Printf(color.HiRedString("[lost-heartbeat] There was an error with the initial query: %v", err.Error()))
		return nil, err
	}
	if respCode/100 != 2 {
		msg := fmt.Sprintf("[lost-heartbeat] Non 200 response received from the initial query: %v, %s", respCode, body)
		log.Printf(color.HiRedString(msg))
		return nil, errors.New(msg)

	}
	hrresp := device.HeartbeatLostQueryResponse{}

	err = json.Unmarshal(body, &hrresp)
	if err != nil {
		log.Printf(color.HiRedString("Couldn't unmmarshal response: %v", err.Error()))
		return nil, err
	}

	alerts, err := processResponse(hrresp)
	//process the alerts

	return alerts, err

}
