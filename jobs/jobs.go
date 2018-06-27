package jobs

import "github.com/byuoitav/state-parsing/actions"

type Job interface {
	Run() []actions.ActionPayload
}

var Jobs = map[string]Job{
	//	"heartbeat": &heartbeat.LostHeartbeatAlertFactory{},
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
