package state

import (
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/common/log"
)

var roomStateChan chan State
var deviceStateChan chan State

// State is a representation of an update for an entry in an elk static index
type State struct {
	ID    string      // id of the document to update in elk
	Key   string      // key to update in a entry in the static index
	Value interface{} // value of key to set in static index
}
type Manager struct {
	IngestionChannel chan State
	Ticker           *time.Ticker
	DispatchType     string
	DeltaMap         map[string]map[string]interface{}
}

func startManager(toStart Manager) error {
	toStart.DeltaMap = make(map[string]map[string]interface{})

	for {
		select {
		case state := <-toStart.IngestionChannel:

			queueStateDelta(state, toStart.DeltaMap)

		case <-toStart.Ticker.C:
			log.L.Debugf("Dispatching %v state", toStart.DispatchType)
			go dispatchState(toStart.DeltaMap, toStart.DispatchType)

			toStart.DeltaMap = make(map[string]map[string]interface{})
			log.L.Debugf("Finished dispatching %v state; successfully reset map.", toStart.DispatchType)
		}
	}
}

// StartDistributor sends collected state updates every <interval> to elk static indicies.
func StartDistributor(interval time.Duration) {
	log.L.Infof("[Distributor] Starting")

	roomStateChan = make(chan State, 2500)
	deviceStateChan = make(chan State, 2500)

	roomTicker := time.NewTicker(interval)
	deviceTicker := time.NewTicker(interval)

	wg := sync.WaitGroup{}
	wg.Add(2)

	//room state manager
	go startManager(Manager{
		IngestionChannel: roomStateChan,
		Ticker:           roomTicker,
		DispatchType:     ROOM,
	})
	//device state manager
	go startManager(Manager{
		IngestionChannel: deviceStateChan,
		Ticker:           deviceTicker,
		DispatchType:     DEVICE,
	})

	wg.Wait()
}

func queueStateDelta(state State, mapToUse map[string]map[string]interface{}) {
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
