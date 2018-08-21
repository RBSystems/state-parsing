package state

import (
	"fmt"
	"strings"

	"github.com/byuoitav/common/events"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
)

// DistributeEvent buffers state for an event.
func DistributeEvent(event elkreporting.ElkEvent) {
	if !strings.EqualFold(event.EventTypeString, "CORESTATE") && !strings.EqualFold(event.EventTypeString, "DETAILSTATE") {
		// only distribute corestate/detailstate events (at least for now!)
		return
	}

	// TODO should this just use event.Hostname?
	deviceID := fmt.Sprintf("%s-%s-%s", event.Building, event.Room, event.Event.Event.Device)
	//to transition to use the device hostname, all we'd have to do is:
	//deviceID := event.Event.Event.Device

	roomID := fmt.Sprintf("%s-%s", event.Building, event.Room)

	BufferState(State{
		ID:    deviceID,
		Key:   event.Event.Event.EventInfoKey,
		Value: event.Event.Event.EventInfoValue,
	}, "device")

	// check if it's a userinput event. if so, update the last-user-input field
	if event.EventCauseString == "USERINPUT" {
		BufferState(State{
			ID:    deviceID,
			Key:   "last-user-input",
			Value: event.Timestamp,
		}, "device")

		// update the room last-user-input field as well
		BufferState(State{
			ID:    roomID,
			Key:   "last-user-input",
			Value: event.Timestamp,
		}, "room")
	}

	BufferState(State{
		ID:    roomID,
		Key:   "last-state-received",
		Value: event.Timestamp,
	}, "room")

}

// DistributeHeartbeat buffers state related to a heartbeat event
func DistributeHeartbeat(event events.Event) {
	if event.Event.Type != events.HEARTBEAT {
		// we don't care
		return
	}

	BufferState(State{
		ID:    event.Hostname,
		Key:   "last-heartbeat",
		Value: event.Timestamp,
	}, "device")

	BufferState(State{
		ID:    event.Building + "-" + event.Room,
		Key:   "last-heartbeat-received",
		Value: event.Timestamp,
	}, "room")
}

// BufferState adds/updates a piece of data in the local state map
func BufferState(state State, mapType string) {
	//check for a nil interface
	if state.Value == nil {
		return
	}

	//check for a blank string
	switch state.Value.(type) {
	case string:
		if len(state.Value.(string)) == 0 {
			return
		}
	}

	switch mapType {
	case "room":
		roomStateChan <- state
	case "device":
		deviceStateChan <- state
	}
}
