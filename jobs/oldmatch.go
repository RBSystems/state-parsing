package jobs

import (
	"fmt"
	"regexp"
	"time"

	"github.com/byuoitav/event-translator-microservice/elkreporting"
)

// OldMatchConfig contains the logic for building/matching regex for events that come in
type OldMatchConfig struct {
	Count int

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

func (r *runner) buildOldMatchRegex() MatchConfig {
	m := &OldMatchConfig{}
	m.Count = 0

	// build the regex for each field
	if len(m.Hostname) > 0 {
		m.Regex.Hostname = regexp.MustCompile(m.Hostname)
		m.Count++
	}

	if len(m.Timestamp) > 0 {
		m.Regex.Timestamp = regexp.MustCompile(m.Timestamp)
		m.Count++
	}

	if len(m.LocalEnvironment) > 0 {
		m.Regex.LocalEnvironment = regexp.MustCompile(m.LocalEnvironment)
		m.Count++
	}

	if len(m.Building) > 0 {
		m.Regex.Building = regexp.MustCompile(m.Building)
		m.Count++
	}

	if len(m.Room) > 0 {
		m.Regex.Room = regexp.MustCompile(m.Room)
		m.Count++
	}

	if len(m.Event.Type) > 0 {
		m.Regex.Event.Type = regexp.MustCompile(m.Event.Type)
		m.Count++
	}

	if len(m.Event.Requestor) > 0 {
		m.Regex.Event.Requestor = regexp.MustCompile(m.Event.Requestor)
		m.Count++
	}

	if len(m.Event.EventCause) > 0 {
		m.Regex.Event.EventCause = regexp.MustCompile(m.Event.EventCause)
		m.Count++
	}

	if len(m.Event.Device) > 0 {
		m.Regex.Event.Device = regexp.MustCompile(m.Event.Device)
		m.Count++
	}

	if len(m.Event.EventInfoKey) > 0 {
		m.Regex.Event.EventInfoKey = regexp.MustCompile(m.Event.EventInfoKey)
		m.Count++
	}

	if len(m.Event.EventInfoValue) > 0 {
		m.Regex.Event.EventInfoValue = regexp.MustCompile(m.Event.EventInfoValue)
		m.Count++
	}

	return m
}

func (m *OldMatchConfig) doesEventMatch(e interface{}) bool {
	event, ok := e.(*elkreporting.ElkEvent)
	if !ok {
		return false
	}

	if m.Count == 0 {
		return true
	}

	if m.Regex.Hostname != nil {
		reg := m.Regex.Hostname.Copy()
		if !reg.MatchString(event.Hostname) {
			return false
		}
	}

	if m.Regex.Timestamp != nil {
		reg := m.Regex.Timestamp.Copy()
		if !reg.MatchString(event.Timestamp.Format(time.RFC3339)) {
			return false
		}
	}

	if m.Regex.LocalEnvironment != nil {
		reg := m.Regex.LocalEnvironment.Copy()
		if !reg.MatchString(fmt.Sprintf("%v", event.LocalEnvironment)) {
			return false
		}
	}

	if m.Regex.Building != nil {
		reg := m.Regex.Building.Copy()
		if !reg.MatchString(event.Building) {
			return false
		}
	}

	if m.Regex.Room != nil {
		reg := m.Regex.Room.Copy()
		if !reg.MatchString(event.Room) {
			return false
		}
	}

	if m.Regex.Event.Type != nil {
		reg := m.Regex.Event.Type.Copy()
		if !reg.MatchString(fmt.Sprintf("%v", event.Event.Event.Type)) {
			return false
		}
	}

	if m.Regex.Event.Requestor != nil {
		reg := m.Regex.Event.Requestor.Copy()
		if !reg.MatchString(event.Event.Event.Requestor) {
			return false
		}
	}

	if m.Regex.Event.EventCause != nil {
		reg := m.Regex.Event.EventCause.Copy()
		if !reg.MatchString(fmt.Sprintf("%v", event.Event.Event.EventCause)) {
			return false
		}
	}

	if m.Regex.Event.Device != nil {
		reg := m.Regex.Event.Device.Copy()
		if !reg.MatchString(event.Event.Event.Device) {
			return false
		}
	}

	if m.Regex.Event.EventInfoKey != nil {
		reg := m.Regex.Event.EventInfoKey.Copy()
		if !reg.MatchString(event.Event.Event.EventInfoKey) {
			return false
		}
	}

	if m.Regex.Event.EventInfoValue != nil {
		reg := m.Regex.Event.EventInfoValue.Copy()
		if !reg.MatchString(event.Event.Event.EventInfoValue) {
			return false
		}
	}

	return true
}
