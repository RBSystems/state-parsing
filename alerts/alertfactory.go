package alerts

import (
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/state-parsing/alerts/base"
	"github.com/byuoitav/state-parsing/alerts/device"
	"github.com/byuoitav/state-parsing/notifications"
)

type AlertFactory struct {
	AlertsToSend map[string][]base.Alert
}

func (a *AlertFactory) Pre() (error, bool) {
	a.AlertsToSend = make(map[string][]base.Alert)
	return nil, true
}

func (a *AlertFactory) Post(err error) {
	// ignore the error, try to send things anyways

	engines := notifications.GetNotificationEngines()
	reports := []base.AlertReport{}

	log.L.Infof("Sending notifications...")

	for k, v := range a.AlertsToSend {
		reps, err := engines[k].SendNotifications(v)
		if err != nil {
			log.L.Errorf("issue sending the %v notifications. error: %s", k, err)
		}

		reports = append(reports, reps...)
	}

	log.L.Infof("Marking alert as sent.")
	device.MarkLastAlertSent(reports)
}
