package jobs

import "github.com/byuoitav/state-parsing/actions"

type Job interface {
	Run() []actions.Action // this type needs to be changed
}

var Jobs = map[string]Job{
	//	"heartbeat": &heartbeat.LostHeartbeatAlertFactory{},
}
