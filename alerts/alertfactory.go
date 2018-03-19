package alerts

import (
	"github.com/byuoitav/state-parsing/alerts/base"
	"github.com/byuoitav/state-parsing/alerts/device"
	"github.com/byuoitav/state-parsing/tasks/task"
)

type AlertFactory struct {
	task.Task

	AlertsToSend map[string][]base.Alert
}

func (a *AlertFactory) Pre() (error, bool) {
	a.AlertsToSend = make(map[string][]base.Alert)
	return nil, true
}

func (a *AlertFactory) Post(err error) {
	// ignore the error, try to send things anyways

	engines := GetNotificationEngines()
	reports := []base.AlertReport{}

	a.I("Sending notifications...")

	for k, v := range a.AlertsToSend {
		reps, err := engines[k].SendNotifications(v)
		if err != nil {
			a.E("issue sending the %v notifications. error: %s", k, err)
		}

		reports = append(reports, reps...)
	}

	a.I("Marking alert as sent.")

	device.MarkLastAlertSent(reports)
}
