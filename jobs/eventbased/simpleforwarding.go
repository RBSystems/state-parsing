package eventbased

import (
	"os"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/byuoitav/state-parser/actions/action"
	"github.com/byuoitav/state-parser/elk"
	"github.com/byuoitav/state-parser/state"
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
}

// SimpleForwarding is the name of this job
const SimpleForwarding = "simple-forwarding"

// SimpleForwardingJob is exported to add it as a job.
type SimpleForwardingJob struct {
}

// Run fowards events to an elk timeseries index.
func (*SimpleForwardingJob) Run(context interface{}) []action.Payload {

	//	cache.GetCache(cache.DEFAULT)

	switch v := context.(type) {
	case *elkreporting.ElkEvent:
		go state.Forward(*v, elk.UpdateHeader{
			Index: elk.GenerateIndexName(elk.OIT_AV_ALL),
			Type:  "oitaveventprd",
		})

		state.DistributeEvent(*v)
	case elkreporting.ElkEvent:
		go state.Forward(v, elk.UpdateHeader{
			Index: elk.GenerateIndexName(elk.OIT_AV_ALL),
			Type:  "oitaveventprd",
		})

		state.DistributeEvent(v)
	default:
	}

	return nil
}
