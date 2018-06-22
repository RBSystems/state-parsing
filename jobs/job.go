package jobs

type Job interface {
	Run() []Action // this type needs to be changed
}

type Action interface {
	Execute()
}

var Jobs = map[string]Job{
	//	"heartbeat": &heartbeat.LostHeartbeatAlertFactory{},
}

var Actions = map[string]Action{
	// fill actions in here
	//	actions.MOM: mom.MomNotificationEngine{},
}
