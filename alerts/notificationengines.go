package alerts

import (
	"github.com/byuoitav/state-parsing/alerts/base"
	"github.com/byuoitav/state-parsing/alerts/slacknotifications"
)

type NotificationEngine interface {
	SendNotifications([]base.Alert) ([]base.AlertReport, error) //I'm not sure still if we care about AlertReports yet, but easier to add it now than later.
}

func GetNotificationEngines() map[string]NotificationEngine {
	toReturn := make(map[string]NotificationEngine)

	toReturn[base.SLACK] = &slacknotifications.SlackNotificationEngine{ChannelIdentifier: "/T0311JJTE/B6ZMQBZ2B/dwVLO10Iu03mJ9IqmPExGoy3"}

	return toReturn
}
