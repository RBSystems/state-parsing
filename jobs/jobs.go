package jobs

import (
	"github.com/byuoitav/state-parser/actions/action"
	"github.com/byuoitav/state-parser/jobs/actiongen"
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

// JobConfig .
type JobConfig struct {
	Name     string           `json:"name"`
	Triggers []Trigger        `json:"triggers"`
	Enabled  bool             `json:"enabled"`
	Action   actiongen.Config `json:"action"`
}

// Trigger .
type Trigger struct {
	Type     string          `json:"type"`      // required for all
	At       string          `json:"at"`        // required for 'time'
	Every    string          `json:"every"`     // required for 'interval'
	NewMatch *NewMatchConfig `json:"new-match"` // required for 'event'
	OldMatch *OldMatchConfig `json:"old-match"` // required for 'event'
}
