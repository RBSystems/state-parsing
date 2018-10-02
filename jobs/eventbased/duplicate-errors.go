package eventbased

import (
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/state-parser/actions/action"
	"github.com/byuoitav/state-parser/jobs/eventstore"
)

// DuplicateErrorsJob will watch for "error" events, and if it sees a lot in a short time period, it will fire an alert.
type DuplicateErrorsJob struct {
}

const maxAlertCount = 3

type key struct {
	GeneratingSystem string
	EventKey         string
	EventValue       string
}

var eventStore *eventstore.Store

// build event store in init
func init() {
	eventStore = eventstore.New(areThereDuplicateErrors)
}

// Run is executed each time an event comes through
func (*DuplicateErrorsJob) Run(context interface{}, actionWrite chan action.Payload) {
	// validate that context contains the correct type
	event, ok := context.(*events.Event)
	if !ok {
		log.L.Warnf("DuplicateErrorsJob only works with v2 events.")
	}

	// build key for store
	k := key{
		GeneratingSystem: event.GeneratingSystem,
		EventKey:         event.Key,
		EventValue:       event.Value,
	}

	eventStore.Store(k, *event)
	/*
		// send the alerts
		// clear the key from the map
		seen.Lock()
		delete(seen.at, k)
		seen.Unlock()

		// fire off alerts
		actionWrite <- action.Payload{
			Type:   actions.Slack,
			Device: event.TargetDevice.DeviceID,
			Content: slack.Attachment{
				Fallback: fmt.Sprintf("Duplicate errors detected on %v", event.TargetDevice.DeviceID),
				Title:    "Error",
				Fields: []slack.AlertField{
					slack.AlertField{
						Title: "Key",
						Value: event.Key,
						Short: true,
					},
					slack.AlertField{
						Title: "Value",
						Value: event.Value,
						Short: true,
					},
				},
				Color: "danger",
			},
		}
	*/
}

func areThereDuplicateErrors(events []events.Event) {
}
