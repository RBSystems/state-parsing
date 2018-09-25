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
	"heartbeat-lost":         &heartbeat.LostJob{},
	"heartbeat-restored":     &heartbeat.RestoredJob{},
	"room-update":            &timebased.RoomUpdateJob{},
	"general-alert-clearing": &timebased.GeneralAlertClearingJob{},
	"simple-forwarding":      &eventbased.SimpleForwardingJob{},
	"duplicate-errors":       &eventbased.DuplicateErrorsJob{},
}

// JobConfig .
type JobConfig struct {
	Name     string    `json:"name"`
	Triggers []Trigger `json:"triggers"`
	Enabled  bool      `json:"enabled"`
}

// Trigger .
type Trigger struct {
	Type     string          `json:"type"`      // required for all
	At       string          `json:"at"`        // required for 'time'
	Every    string          `json:"every"`     // required for 'interval'
	NewMatch *NewMatchConfig `json:"new-match"` // required for 'event'
	OldMatch *OldMatchConfig `json:"old-match"` // required for 'event'
}
