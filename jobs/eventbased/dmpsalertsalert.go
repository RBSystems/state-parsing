package eventbased

import (
	"fmt"
	"time"

	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/byuoitav/state-parser/actions"
	"github.com/byuoitav/state-parser/actions/action"
	"github.com/byuoitav/state-parser/actions/mom"
)

// var (
// 	// SentinelEndpoint is the url to send alerts to
// 	SentinelEndpoint = os.Getenv("SENTINEL_ENDPOINT")
// )

func init() {
	// if len(APIForward) == 0 || len(HeartbeatForward) == 0 {
	// 	log.L.Fatalf("$ELASTIC_API_EVENTS and $ELASTIC_HEARTBEAT_EVENTS must be set.")
	// }
	// log.L.Infof("\n\nForwarding URLs:\n\tAPI Forward:\t\t%v\n\tSecond API Forward\t\t%v\n\tHeartbeat Forward:\t%v\n", APIForward, SecondAPIForward, HeartbeatForward)

	// if len(DMPSEventsForward) == 0 || len(DMPSHeartbeatForward) == 0 {
	// 	log.L.Fatalf("$ELASTIC_DMPS_EVENTS and $ELASTIC_DMPS_HEARTBEATS must be set.")
	// }
}

// DMPSAlertsAlert is the name of this job
const DMPSAlertsAlert = "dmps-alerts-alert"

// DMPSAlertsAlertJob is exported to add it as a job.
type DMPSAlertsAlertJob struct {
}

// Run fowards events to an elk timeseries index.
func (*DMPSAlertsAlertJob) Run(context interface{}, parameter string) []action.Payload {

	var theEvent elkreporting.ElkEvent

	switch context.(type) {
	case *elkreporting.ElkEvent:
		theEvent = context.(elkreporting.ElkEvent)
	case elkreporting.ElkEvent:
		theEvent = context.(elkreporting.ElkEvent)
	default:
	}

	timeParse, _ := time.Parse(time.RFC3339, theEvent.Timestamp)
	timeOutput := timeParse.Format("01/02/2006 15:04:05")

	toReturn := []action.Payload{
		action.Payload{
			Type:   actions.Mom,
			Device: theEvent.Hostname,
			Content: mom.Alert{
				Host:     theEvent.Hostname,
				Element:  theEvent.Event.Event.Device,
				Severity: 2, //warning,
				AlertOutput: fmt.Sprintf("DMPS [%v] Device [%v] alerting on [%v] with value [%v",
					theEvent.Hostname, theEvent.Event.Event.Device, theEvent.Event.Event.EventInfoKey, theEvent.Event.Event.EventInfoValue),
				AlertTime: timeOutput,
				Service:   "", // hard coded by the Sentinel/MOM group
				KB:        "KB0000000",
				Ticket:    "",
			},
		},
	}

	return toReturn
}
