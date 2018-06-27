package actions

const (
	MOM   = "mom"
	SLACK = "slack"
)

type Action interface {
	Execute(payload ActionPayload)
}

type ActionPayload struct {
	ActionName string // name of the alert, found in constants above
	Device     string // the device the alert corresponds to
	Content    interface{}
}

var Actions = map[string]Action{
	// fill actions in here
	//	MOM: mom.MomNotificationEngine{},
}
