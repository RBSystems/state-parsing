package jobs

import (
	"fmt"
	"regexp"

	"github.com/byuoitav/event-translator-microservice/elkreporting"
)

type MatchConfig struct {
	Hostname         string `json:"hostname,omitempty"`
	Timestamp        string `json:"timestamp,omitempty"`
	LocalEnvironment string `json:"localEnvironment,omitempty"`
	Building         string `json:"building,omitempty"`
	Room             string `json:"room,omitempty"`

	Event struct {
		Type           string `json:"type,omitempty"`
		Requestor      string `json:"requestor,omitempty"`
		EventCause     string `json:"eventCause,omitempty"`
		Device         string `json:"device,omitempty"`
		EventInfoKey   string `json:"eventInfoKey,omitempty"`
		EventInfoValue string `json:"eventInfoValue,omitempty"`
	} `json:"event,omitempty"`

	Regex struct {
		Hostname         *regexp.Regexp
		Timestamp        *regexp.Regexp
		LocalEnvironment *regexp.Regexp
		Building         *regexp.Regexp
		Room             *regexp.Regexp

		Event struct {
			Type           *regexp.Regexp
			Requestor      *regexp.Regexp
			EventCause     *regexp.Regexp
			Device         *regexp.Regexp
			EventInfoKey   *regexp.Regexp
			EventInfoValue *regexp.Regexp
		}
	}
}

func (r *runner) buildMatchRegex() {
	// build the regex for each field
	// TODO validate at least one regex is created
	if len(r.Trigger.Match.Hostname) > 0 {
		r.Trigger.Match.Regex.Hostname = regexp.MustCompile(r.Trigger.Match.Hostname)
	}

	if len(r.Trigger.Match.Timestamp) > 0 {
		r.Trigger.Match.Regex.Timestamp = regexp.MustCompile(r.Trigger.Match.Timestamp)
	}

	if len(r.Trigger.Match.LocalEnvironment) > 0 {
		r.Trigger.Match.Regex.LocalEnvironment = regexp.MustCompile(r.Trigger.Match.LocalEnvironment)
	}

	if len(r.Trigger.Match.Building) > 0 {
		r.Trigger.Match.Regex.Building = regexp.MustCompile(r.Trigger.Match.Building)
	}

	if len(r.Trigger.Match.Room) > 0 {
		r.Trigger.Match.Regex.Room = regexp.MustCompile(r.Trigger.Match.Room)
	}

	if len(r.Trigger.Match.Event.Type) > 0 {
		r.Trigger.Match.Regex.Event.Type = regexp.MustCompile(r.Trigger.Match.Event.Type)
	}

	if len(r.Trigger.Match.Event.Requestor) > 0 {
		r.Trigger.Match.Regex.Event.Requestor = regexp.MustCompile(r.Trigger.Match.Event.Requestor)
	}

	if len(r.Trigger.Match.Event.EventCause) > 0 {
		r.Trigger.Match.Regex.Event.EventCause = regexp.MustCompile(r.Trigger.Match.Event.EventCause)
	}

	if len(r.Trigger.Match.Event.Device) > 0 {
		r.Trigger.Match.Regex.Event.Device = regexp.MustCompile(r.Trigger.Match.Event.Device)
	}

	if len(r.Trigger.Match.Event.EventInfoKey) > 0 {
		r.Trigger.Match.Regex.Event.EventInfoKey = regexp.MustCompile(r.Trigger.Match.Event.EventInfoKey)
	}

	if len(r.Trigger.Match.Event.EventInfoValue) > 0 {
		r.Trigger.Match.Regex.Event.EventInfoValue = regexp.MustCompile(r.Trigger.Match.Event.EventInfoValue)
	}
}

func (r *runner) doesEventMatch(event elkreporting.ElkEvent) bool {
	if r.Trigger.Match.Regex.Hostname != nil {
		passed := r.Trigger.Match.Regex.Hostname.MatchString(event.Hostname)
		if !passed {
			return false
		}
	}

	if r.Trigger.Match.Regex.Timestamp != nil {
		passed := r.Trigger.Match.Regex.Timestamp.MatchString(event.Timestamp)
		if !passed {
			return false
		}
	}

	if r.Trigger.Match.Regex.LocalEnvironment != nil {
		passed := r.Trigger.Match.Regex.LocalEnvironment.MatchString(fmt.Sprintf("%v", event.LocalEnvironment))
		if !passed {
			return false
		}
	}

	if r.Trigger.Match.Regex.Building != nil {
		passed := r.Trigger.Match.Regex.Building.MatchString(event.Building)
		if !passed {
			return false
		}
	}

	if r.Trigger.Match.Regex.Room != nil {
		passed := r.Trigger.Match.Regex.Room.MatchString(event.Room)
		if !passed {
			return false
		}
	}

	if r.Trigger.Match.Regex.Event.Type != nil {
		passed := r.Trigger.Match.Regex.Event.Type.MatchString(fmt.Sprintf("%v", event.Event.Event.Type))
		if !passed {
			return false
		}
	}

	if r.Trigger.Match.Regex.Event.Requestor != nil {
		passed := r.Trigger.Match.Regex.Event.Requestor.MatchString(event.Event.Event.Requestor)
		if !passed {
			return false
		}
	}

	if r.Trigger.Match.Regex.Event.EventCause != nil {
		passed := r.Trigger.Match.Regex.Event.EventCause.MatchString(fmt.Sprintf("%v", event.Event.Event.EventCause))
		if !passed {
			return false
		}
	}

	if r.Trigger.Match.Regex.Event.Device != nil {
		passed := r.Trigger.Match.Regex.Event.Device.MatchString(event.Event.Event.Device)
		if !passed {
			return false
		}
	}

	if r.Trigger.Match.Regex.Event.EventInfoKey != nil {
		passed := r.Trigger.Match.Regex.Event.EventInfoKey.MatchString(event.Event.Event.EventInfoKey)
		if !passed {
			return false
		}
	}

	if r.Trigger.Match.Regex.Event.EventInfoValue != nil {
		passed := r.Trigger.Match.Regex.Event.EventInfoValue.MatchString(event.Event.Event.EventInfoValue)
		if !passed {
			return false
		}
	}

	return true
}
