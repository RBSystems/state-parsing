package eventbased

import (
	"os"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/byuoitav/state-parser/actions/action"
	"github.com/byuoitav/state-parser/forwarding"
)

var (
	// APIForward is the url to forward events to
	APIForward = os.Getenv("ELASTIC_API_EVENTS")

	// HeartbeatForward is the url to forward heartbeats to
	HeartbeatForward = os.Getenv("ELASTIC_HEARTBEAT_EVENTS")
)

func init() {
	if len(APIForward) == 0 || len(HeartbeatForward) == 0 {
		log.L.Fatalf("$ELASTIC_API_EVENTS and $ELASTIC_HEARTBEAT_EVENTS must be set.")
	}
	log.L.Infof("\n\nForwarding URLs:\n\tAPI Forward:\t\t%v\n\tHeartbeat Forward:\t%v\n", APIForward, HeartbeatForward)
}

const SimpleForwarding = "simple-forwarding"

type SimpleForwardingJob struct {
}

func (*SimpleForwardingJob) Run(context interface{}) []action.Payload {
	//	var payloads []action.Payload

	switch v := context.(type) {
	case *elkreporting.ElkEvent:
		forwarding.DistributeEvent(*v)
		go forwarding.Forward(*v, APIForward)
	case elkreporting.ElkEvent:
		forwarding.DistributeEvent(v)
		go forwarding.Forward(v, APIForward)
	default:
	}

	return nil
}
