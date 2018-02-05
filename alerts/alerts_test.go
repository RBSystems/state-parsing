package alerts

import (
	"log"
	"testing"

	"github.com/byuoitav/state-parsing/alerts/base"
	"github.com/byuoitav/state-parsing/alerts/device"
	"github.com/byuoitav/state-parsing/alerts/heartbeat"
	"github.com/byuoitav/state-parsing/eventforwarding"
	"github.com/fatih/color"
)

func TestStuff(t *testing.T) {

	go eventforwarding.StartDistributor()

	ha := heartbeat.HeartbeatAlertFactory{}
	alerts, err := ha.Run(1)
	if err != nil {
		log.Printf("error: %v", err.Error())
	}

	reports := []base.AlertReport{}
	engines := GetNotificationEngines()

	for k, v := range alerts {
		reps, err := engines[k].SendNotifications(v)
		if err != nil {
			log.Printf(color.HiRedString("Issue sending the %v notifications. Error: %v", k, err.Error()))
		}
		reports = append(reports, reps...)
	}

	//now we mark the reports as sent
	device.MarkLastAlertSent(reports)

	eventforwarding.StartTicker(3000)
}
