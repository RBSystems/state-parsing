package eventbased

import (
	"os"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/byuoitav/state-parser/actions/action"
	"github.com/byuoitav/state-parser/elk"
	"github.com/byuoitav/state-parser/forwarding"
)

var (
	// APIForward is the url to forward events to
	APIForward = os.Getenv("ELASTIC_API_EVENTS")

	// SecondAPIForward is a second location to forward events to
	SecondAPIForward = os.Getenv("ELASTIC_API_EVENTS_TWO")

	// HeartbeatForward is the url to forward heartbeats to
	HeartbeatForward = os.Getenv("ELASTIC_HEARTBEAT_EVENTS")

	// DMPSEventsForward is the url to forward DMPS events to
	DMPSEventsForward = os.Getenv("ELASTIC_DMPS_EVENTS")

	// DMPSHeartbeatForward is the url to forward DMPS events to
	DMPSHeartbeatForward = os.Getenv("ELASTIC_DMPS_HEARTBEATS")
)

func init() {
	if len(APIForward) == 0 || len(HeartbeatForward) == 0 {
		log.L.Fatalf("$ELASTIC_API_EVENTS and $ELASTIC_HEARTBEAT_EVENTS must be set.")
	}
	log.L.Infof("\n\nForwarding URLs:\n\tAPI Forward:\t\t%v\n\tSecond API Forward\t\t%v\n\tHeartbeat Forward:\t%v\n", APIForward, SecondAPIForward, HeartbeatForward)

	if len(DMPSEventsForward) == 0 || len(DMPSHeartbeatForward) == 0 {
		log.L.Fatalf("$ELASTIC_DMPS_EVENTS and $ELASTIC_DMPS_HEARTBEATS must be set.")
	}
}

// SimpleForwarding is the name of this job
const SimpleForwarding = "simple-forwarding"

// SimpleForwardingJob is exported to add it as a job.
type SimpleForwardingJob struct {
}

// Run fowards events to an elk timeseries index.
func (*SimpleForwardingJob) Run(context interface{}) []action.Payload {
	switch v := context.(type) {
	case *elkreporting.ElkEvent:
		forwarding.DistributeEvent(*v)
		go forwarding.Forward(*v, elk.UpdateHeader{
			Index: elk.GenerateIndexName(elk.OIT_AV),
			Type:  "oitavprdevent",
		})
	case elkreporting.ElkEvent:
		forwarding.DistributeEvent(v)
		go forwarding.Forward(v, elk.UpdateHeader{
			Index: elk.GenerateIndexName(elk.OIT_AV),
			Type:  "oitavprdevent",
		})
	default:
	}

	return nil
}
