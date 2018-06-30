package actions

import (
	"os"

	"github.com/byuoitav/state-parsing/actions/action"
	"github.com/byuoitav/state-parsing/actions/slack"
)

const (
	// action types
	MOM   = "mom"
	SLACK = "slack"
)

type Action interface {
	Execute(a action.Action) action.Result
}

var Actions = map[string]Action{
	// fill actions in here
	SLACK: &slack.SlackAction{ChannelIdentifier: os.Getenv("SLACK_HEARTBEAT_CHANNEL")},
	//	MOM: mom.MomNotificationEngine{},
}
