package eventbased

import (
	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/byuoitav/state-parser/actions/action"
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

// DMPSAlerts is the name of this job
const DMPSAlertsAlert = "dmps-alerts-alert"

// DMPSAlertsJob is exported to add it as a job.
type DMPSAlertsJob struct {
}

// Run fowards events to an elk timeseries index.
func (*DMPSAlertsJob) Run(context interface{}) []action.Payload {

	var theEvent elkreporting.ElkEvent

	switch v := context.(type) {
	case *elkreporting.ElkEvent:
		theEvent = &context
	case elkreporting.ElkEvent:
		theEvent = context
	default:
	}

	toReturn := []action.Payload {

	};

	return nil
}
