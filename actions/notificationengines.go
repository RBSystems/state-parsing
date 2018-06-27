package actions

import (
	"os"

	"github.com/byuoitav/state-parsing/actions/mom"
	"github.com/byuoitav/state-parsing/actions/slack"
	"github.com/byuoitav/state-parsing/alerts/base"
)

type NotificationEngine interface {
	SendNotifications([]base.Alert) ([]base.AlertReport, error) //I'm not sure still if we care about AlertReports yet, but easier to add it now than later.
}

func GetNotificationEngines() map[string]NotificationEngine {
	toReturn := make(map[string]NotificationEngine)

	toReturn[SLACK] = &slack.SlackNotificationEngine{ChannelIdentifier: os.Getenv("SLACK_HEARTBEAT_CHANNEL")}
	toReturn[MOM] = &mom.MomNotificationEngine{}

	return toReturn
}
