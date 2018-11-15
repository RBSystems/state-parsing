package jobs

import (
	"github.com/byuoitav/state-parser/actions/action"
	"github.com/byuoitav/state-parser/jobs/eventbased"
	"github.com/byuoitav/state-parser/jobs/timebased"
	"github.com/byuoitav/state-parser/jobs/timebased/heartbeat"
)

// Job .
type Job interface {
	Run(ctx interface{}, actionWrite chan action.Payload)
}

// Jobs .
var Jobs = map[string]Job{
	heartbeat.HeartbeatLost:        &heartbeat.LostJob{},
	heartbeat.HeartbeatRestored:    &heartbeat.RestoredJob{},
	timebased.RoomUpdate:           &timebased.RoomUpdateJob{},
	timebased.GeneralAlertClearing: &timebased.GeneralAlertClearingJob{},
	eventbased.SimpleForwarding:    &eventbased.SimpleForwardingJob{},
}
