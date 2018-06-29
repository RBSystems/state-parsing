package device

import (
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/state-parsing/forwarding"
)

//toMark is the list of rooms, There may be one or more of them
//secondaryAlertType is the type of alert marking as (e.g. heartbeat)
//secondaryAlertData is the data to be filled there (e.g. last-heartbeat-received, etc)
func MarkAsAlerting(toMark []string, secondaryAlertType string, secondaryAlertData map[string]interface{}) {
	//build our general alerting
	alerting := forwarding.StateDistribution{
		Key:   "alerting",
		Value: true,
	}

	secondaryAlertValue := make(map[string]interface{})
	secondaryAlertValue[secondaryAlertType] = secondaryAlertData

	//bulid our specifc alert
	secondaryAlert := forwarding.StateDistribution{
		Key:   "alerts",
		Value: secondaryAlertValue,
	}

	//ship it off to go with the rest
	for i := range toMark {
		log.L.Infof("Marking %s as alerting", toMark[i])
		forwarding.SendToStateBuffer(alerting, toMark[i], "device")
		forwarding.SendToStateBuffer(secondaryAlert, toMark[i], "device")
	}
}

/*
func MarkLastAlertSent(reps []base.AlertReport, secondaryAlertType string) error {
	//for now we assume that we only send device-based alerts
	for r := range reps {
		if !reps[r].Success {
			//we didn't actually get the notification sent
			continue
		}

		tertiary := make(map[string]interface{})
		tertiary["alert-sent"] = reps[r].Message

		secondary := make(map[string]interface{})
		secondary[heartbeat.LOST_HEARTBEAT] = tertiary

		forwarding.SendToStateBuffer(forwarding.StateDistribution{
			Key:   "alerts",
			Value: secondary,
		},
			reps[r].Device, "device")
	}
	return nil
}

func MarkDevicesAsNotAlerting(deviceIDs []string) {
	secondaryData := make(map[string]map[string]interface{})
	secondaryData[heartbeat.LOST_HEARTBEAT] = make(map[string]interface{})

	secondaryData[heartbeat.LOST_HEARTBEAT]["alerting"] = false
	secondaryData[heartbeat.LOST_HEARTBEAT]["message"] = fmt.Sprintf("Alert cleared at %s", time.Now().Format(time.RFC3339))

	secondaryStatus := forwarding.StateDistribution{
		Key:   "alerts",
		Value: secondaryData,
	}

	alertingStatus := forwarding.StateDistribution{
		Key:   "alerting",
		Value: false,
	}

	for _, id := range deviceIDs {
		log.L.Info("Marking %s as not alerting", id)
		forwarding.SendToStateBuffer(secondaryStatus, id, "device")
		forwarding.SendToStateBuffer(alertingStatus, id, "device")
	}
}
*/
