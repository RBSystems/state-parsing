package forwarding

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/common/events"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
)

// State is a representation of an update for an entry in an elk static index
type State struct {
	ID    string      // id of the document to update in elk
	Key   string      // key to update in a entry in the static index
	Value interface{} // value of key to set in static index
}

var roomStateChan chan State
var deviceStateChan chan State

// StartDistributor sends collected state updates every <interval> to elk static indicies.
func StartDistributor(interval time.Duration) {
	log.L.Infof("[Distributor] Starting")

	roomStateChan = make(chan State, 2500)
	deviceStateChan = make(chan State, 2500)

	roomStateMap := make(map[string]map[string]interface{})
	deviceStateMap := make(map[string]map[string]interface{})

	roomTicker := time.NewTicker(interval)
	deviceTicker := time.NewTicker(interval)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		for {
			select {
			case state := <-roomStateChan:
				bufferState(state, roomStateMap)
			case <-roomTicker.C:
				log.L.Debugf("Tick; Dispatching room state")

				go dispatchState(roomStateMap, "room")
				roomStateMap = make(map[string]map[string]interface{})

				log.L.Debugf("Finished dispatching room state; successfully reset map.")
			}
		}
	}()

	go func() {
		for {
			select {
			case state := <-deviceStateChan:
				bufferState(state, deviceStateMap)
			case <-deviceTicker.C:
				log.L.Debugf("Tick; Dispatching device state")

				go dispatchState(deviceStateMap, "device")
				deviceStateMap = make(map[string]map[string]interface{})

				log.L.Debugf("Finished dispatching device state; successfully reset map.")
			}
		}
	}()

	wg.Wait()
}

// DistributeEvent buffers state for an event.
func DistributeEvent(event elkreporting.ElkEvent) {
	if !strings.EqualFold(event.EventTypeString, "CORESTATE") && !strings.EqualFold(event.EventTypeString, "DETAILSTATE") {
		// only distribute corestate/detailstate events (at least for now!)
		return
	}

	// TODO should this just use event.Hostname?
	deviceID := fmt.Sprintf("%s-%s-%s", event.Building, event.Room, event.Event.Event.Device)
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

	// TODO we need to update the room state
	// we need to mark the room to be cheked and updated at the next roomTick
	// roomUpdateChan <- event.Building + "-" + event.Room
}

// DistributeHeartbeat buffers state related to a heartbeat event
func DistributeHeartbeat(event events.Event) {
	if event.Event.Type == events.HEARTBEAT {
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

func bufferState(state State, mapToUse map[string]map[string]interface{}) {
	if _, ok := mapToUse[state.ID]; ok {
		// check to make sure that Value is a Map
		if _, ok := state.Value.(map[string]interface{}); !ok {
			mapToUse[state.ID][state.Key] = state.Value
			return
		}

		// make sure it even exists
		if _, ok := mapToUse[state.ID][state.Key]; !ok {
			mapToUse[state.ID][state.Key] = state.Value
			return
		}

		if _, ok := mapToUse[state.ID][state.Key].(map[string]interface{}); !ok {
			mapToUse[state.ID][state.Key] = state.Value
			return
		}

		// state.Value is also a map
		var a map[string]interface{}
		a = state.Value.(map[string]interface{})
		var b map[string]interface{}
		b = mapToUse[state.ID][state.Key].(map[string]interface{})

		// now we get to compare the child values
		replaceMapValues(&a, &b)
		return
	}

	mapToUse[state.ID] = make(map[string]interface{})
	mapToUse[state.ID][state.Key] = state.Value
}

// will copy map a to b, adding values and overwriting values as found
func replaceMapValues(a, b *map[string]interface{}) {
	for k, v := range *a {
		if strings.Contains(reflect.TypeOf(v).String(), "map[string]interface") {
			bval, ok := (*b)[k]

			//it doesn't exist, just copy it
			if !ok {
				(*b)[k] = v
				continue
			}

			//it exists, check to see if it's a map
			if !strings.Contains(reflect.TypeOf(bval).String(), "map[string]interface") {
				//it's not, replace
				(*b)[k] = v
				continue
			}

			_a := v.(map[string]interface{})
			_b := (*b)[k].(map[string]interface{})

			//it is, so we need to recuse into it
			replaceMapValues(&_a, &_b)
		}
		(*b)[k] = v
	}
}
