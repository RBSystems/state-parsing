package jobs

import (
	"github.com/byuoitav/state-parsing/actions/action"
	"github.com/byuoitav/state-parsing/jobs/timebased"
	"github.com/byuoitav/state-parsing/jobs/timebased/heartbeat"
)

type Job interface {
	Run(ctx interface{}) []action.Payload
}

var Jobs = map[string]Job{
	heartbeat.HEARTBEAT_LOST:         &heartbeat.HeartbeatLostJob{},
	heartbeat.HEARTBEAT_RESTORED:     &heartbeat.HeartbeatRestoredJob{},
	timebased.ROOM_UPDATE:            &timebased.RoomUpdateJob{},
	timebased.GENERAL_ALERT_CLEARING: &timebased.GeneralAlertClearingJob{},
}

type JobConfig struct {
	Name     string    `json:"name"`
	Triggers []Trigger `json:"triggers"`
	Enabled  bool      `json:"enabled"`
}

type Trigger struct {
	Type  string      `json:"type"`  // required for all
	At    string      `json:"at"`    // required for 'time'
	Every string      `json:"every"` // required for 'interval'
	Match MatchConfig `json:"match"` // required for 'event'
}
