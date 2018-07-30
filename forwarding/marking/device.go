package marking

import (
	"fmt"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/state-parser/forwarding"
)

// toMark is the list of rooms, There may be one or more of them
// secondaryAlertType is the type of alert marking as (e.g. heartbeat)
// secondaryAlertData is the data to be filled there (e.g. last-heartbeat-received, etc)
func MarkDevicesAsAlerting(toMark []string, secondaryAlertType string, secondaryAlertData map[string]interface{}) {
	//build our general alerting
	alerting := forwarding.State{
		Key:   "alerting",
		Value: true,
	}

	secondaryAlertValue := make(map[string]interface{})
	secondaryAlertValue[secondaryAlertType] = secondaryAlertData

	// bulid our specifc alert
	secondaryAlert := forwarding.State{
		Key:   "alerts",
		Value: secondaryAlertValue,
	}

	// ship it off to go with the rest
	for i := range toMark {
		log.L.Infof("Marking %s as alerting", toMark[i])
		alerting.ID = toMark[i]
		secondaryAlert.ID = toMark[i]

		forwarding.BufferState(alerting, "device")
		forwarding.BufferState(secondaryAlert, "device")
	}
}

func MarkDevicesAsNotHeartbeatAlerting(deviceIDs []string) {
	secondaryData := make(map[string]map[string]interface{})
	secondaryData["lost-heartbeat"] = make(map[string]interface{})

	secondaryData["lost-heartbeat"]["alerting"] = false
	secondaryData["lost-heartbeat"]["message"] = fmt.Sprintf("Alert cleared at %s", time.Now().Format(time.RFC3339))

	secondaryStatus := forwarding.State{
		Key:   "alerts",
		Value: secondaryData,
	}

	alertingStatus := forwarding.State{
		Key:   "alerting",
		Value: false,
	}

	for _, id := range deviceIDs {
		log.L.Info("Marking %s as not alerting", id)

		alertingStatus.ID = id
		secondaryStatus.ID = id

		forwarding.BufferState(secondaryStatus, "device")
		forwarding.BufferState(alertingStatus, "device")
	}
}
