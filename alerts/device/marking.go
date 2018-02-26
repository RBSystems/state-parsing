package device

import (
	"log"

	"github.com/byuoitav/state-parsing/alerts/base"
	"github.com/byuoitav/state-parsing/eventforwarding"
	"github.com/fatih/color"
)

//toMark is the list of rooms, There may be one or more of them
//secondaryAlertType is the type of alert marking as (e.g. heartbeat)
//secondarAlertData is the data to be filled there (e.g. last-heartbeat-received, etc)
func MarkAsAlerting(toMark []string, secondaryAlertType string, secondaryAlertData map[string]interface{}) {

	//build our general alerting
	alerting := eventforwarding.StateDistribution{
		Key:   "alerting",
		Value: true,
	}

	secondaryAlertValue := make(map[string]interface{})
	secondaryAlertValue[secondaryAlertType] = secondaryAlertData

	//bulid our specifc alert
	secondaryAlert := eventforwarding.StateDistribution{
		Key:   "alerts",
		Value: secondaryAlertValue,
	}

	//ship it off to go with the rest
	for i := range toMark {
		log.Printf(color.HiYellowString("Marking as alerting %v", toMark[i]))
		eventforwarding.SendToStateBuffer(alerting, toMark[i], "device")
		eventforwarding.SendToStateBuffer(secondaryAlert, toMark[i], "device")
	}
}

func MarkLastAlertSent(reps []base.AlertReport) error {

	//for now we assume that we only send device-based alerts
	for r := range reps {
		if !reps[r].Success {
			//we didn't actually get the notification sent
			continue
		}

		tertiary := make(map[string]interface{})
		tertiary["alert-sent"] = reps[r].Message

		secondary := make(map[string]interface{})
		secondary[base.LOST_HEARTBEAT] = tertiary

		eventforwarding.SendToStateBuffer(eventforwarding.StateDistribution{
			Key:   "alerts",
			Value: secondary,
		},
			reps[r].Device, "device")
	}
	return nil
}
