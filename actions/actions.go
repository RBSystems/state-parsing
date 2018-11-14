package actions

import (
	"os"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/state-parser/actions/action"
	"github.com/byuoitav/state-parser/actions/email"
	"github.com/byuoitav/state-parser/actions/mom"
	"github.com/byuoitav/state-parser/actions/slack"
)

const (
	// Slack ..
	Slack = "slack"
	// Mom ..
	Mom = "mom"
	//Email ..
	Email = "email"
)

// An Action is a struct that will execute action payloads created by jobs.
type Action interface {
	Execute(a action.Payload) action.Result
}

// Actions is a map of the action name to an actual Action struct.
var Actions = map[string]Action{
	// fill actions in here
	Slack: &slack.Action{ChannelIdentifier: os.Getenv("SLACK_HEARTBEAT_CHANNEL")},
	Mom:   &mom.Action{},
	Email: &email.Action{},
}

func init() {
	slackHeartbeat := os.Getenv("SLACK_HEARTBEAT_CHANNEL")
	if len(slackHeartbeat) == 0 {
		log.L.Fatalf("SLACK_HEARTBEAT_CHANNEL must be set.")
	}
}
