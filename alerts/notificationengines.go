package alerts

import (
	"os"

	"github.com/byuoitav/state-parsing/alerts/base"
	"github.com/byuoitav/state-parsing/alerts/slacknotifications"
)

type NotificationEngine interface {
	SendNotifications([]base.Alert) ([]base.AlertReport, error) //I'm not sure still if we care about AlertReports yet, but easier to add it now than later.
}

func GetNotificationEngines() map[string]NotificationEngine {
	toReturn := make(map[string]NotificationEngine)

	toReturn[base.SLACK] = &slacknotifications.SlackNotificationEngine{ChannelIdentifier: os.Getenv("SLACK_HEARTBEAT_CHANNEL")}

	return toReturn
}
