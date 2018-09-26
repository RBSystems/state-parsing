package eventbased

import (
	"fmt"
	"sync"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/state-parser/actions"
	"github.com/byuoitav/state-parser/actions/action"
	"github.com/byuoitav/state-parser/actions/slack"
)

// DuplicateErrorsJob will watch for "error" events, and if it sees a lot in a short time period, it will fire an alert.
type DuplicateErrorsJob struct {
}

const maxAlertCount = 3

type key struct {
	EventKey   string
	EventValue string
}

type value struct {
	sync.Mutex
	errors []events.Event
}

var initSeen sync.Once
var seen struct {
	sync.RWMutex
	at map[key]*value
}

// Run is executed each time an event comes through
func (*DuplicateErrorsJob) Run(context interface{}, actionWrite chan action.Payload) {
	// initialize seen map
	initSeen.Do(func() {
		seen.at = make(map[key]*value)
	})

	// type check to make sure it's the new event type
	event, ok := context.(*events.Event)
	if !ok {
		log.L.Warnf("DuplicateErrorsJob only works with v2/events.")
		return
	}

	// build key for map
	k := key{
		EventKey:   event.Key,
		EventValue: event.Value,
	}

	seen.Lock()
	seenAt, ok := seen.at[k]
	seen.RUnlock()

	if !ok {
		// it isn't in the map yet, so add it.
		val := &value{}
		val.errors = append(val.errors, *event)

		// there is a *small* chance of losing an event here, but it's highly unlikely.
		seen.Lock()
		seen.at[k] = val
		seen.Unlock()

		return
	}

	seenAt.Lock()
	if len(seenAt.errors) >= maxAlertCount {
		// clear the existing errors, send the alerts
		seenAt.errors = nil
		seenAt.Unlock()
	} else {
		// add error to list
		seenAt.errors = append(seenAt.errors)
		seenAt.Unlock()

		return
	}

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
}
