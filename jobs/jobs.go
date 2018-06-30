package jobs

import (
	"github.com/byuoitav/state-parsing/actions/action"
	"github.com/byuoitav/state-parsing/jobs/timebased/heartbeat"
)

type Job interface {
	Run(context interface{}) []action.Action
}

var Jobs = map[string]Job{
	heartbeat.HEARTBEAT_LOST:     &heartbeat.HeartbeatLostJob{},
	heartbeat.HEARTBEAT_RESTORED: &heartbeat.HeartbeatRestoredJob{},
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
