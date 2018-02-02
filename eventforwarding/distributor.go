package eventforwarding

import (
	"log"
	"reflect"
	"strings"

	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/fatih/color"

	heartbeat "github.com/byuoitav/salt-translator-service/elk"
)

//all we really need to distribute is the event info key/value - where the key is the value to update in the index.
type StateDistribution struct {
	Key   string
	Value interface{}
}

//determine if we run distributed or not
const runLocal bool = true

//this is distribution to the outside areas
var stateCacheMap map[string]chan StateDistribution

//this is for local
var localStateMap map[string]map[string]interface{}

//this is for local
var localRoomStateMap map[string]map[string]interface{}

func StartDistributor() {

	log.Printf("[Distributor] Starting")

	//initialize our ingestion channels
	eventIngestionChannel = make(chan elkreporting.ElkEvent, 1000)
	heartbeatIngestionChannel = make(chan heartbeat.Event, 1000)

	stateCacheMap = make(map[string]chan StateDistribution)
	localStateMap = make(map[string]map[string]interface{})
	localRoomStateMap = make(map[string]map[string]interface{})

	//our loop for ingestion and distribution
	for {

		select {
		case e := <-eventIngestionChannel:
			apiForwardingChannel <- e
			distributeEvent(e)
		case e := <-heartbeatIngestionChannel:
			heartbeatForwardingChannel <- e
			distributeHeartbeat(e)
		case <-localTickerChan:
			//ship it out
			go dispatchLocalState(localStateMap, "device")
			go dispatchLocalState(localRoomStateMap, "room")

			//refresh the maps
			localStateMap = make(map[string]map[string]interface{})
			localRoomStateMap = make(map[string]map[string]interface{})
		}

		//we need to send it on to the ELK stack as-is
	}
}

func distributeEvent(event elkreporting.ElkEvent) {
	log.Printf("[distributor] state recieved")
	if event.EventTypeString != "CORESTATE" && event.EventTypeString != "DETAILSTATE" {
		//we don't care about it for now
		return
	}

	//log.Printf("buildilng event and sending")
	//we need to pull out the values for StateDistributionm
	toSend := StateDistribution{Key: event.Event.Event.EventInfoKey, Value: event.Event.Event.EventInfoValue}

	if runLocal {
		//we need to check if it's a userinput event, if so we need to update the last-user-input field
		localStateBuffering(toSend, event.Building+"-"+event.Room+"-"+event.Event.Event.Device, "device")

		if event.EventCauseString == "USERINPUT" {
			localStateBuffering(StateDistribution{
				Key:   "last-user-input",
				Value: event.Timestamp,
			}, event.Building+"-"+event.Room+"-"+event.Event.Event.Device, "device")

			//we need to update the room as well.
			localStateBuffering(StateDistribution{
				Key:   "last-user-input",
				Value: event.Timestamp,
			}, event.Building+"-"+event.Room, "room")

		}
		localStateBuffering(StateDistribution{
			Key:   "last-state-received",
			Value: event.Timestamp,
		}, event.Building+"-"+event.Room, "room")

		//we need to update the room state

	} else {
		sendToStateBuffering(toSend, event.Building+"-"+event.Room+"-"+event.Event.Event.Device)
		if event.EventCauseString == "USERINPUT" {
			sendToStateBuffering(StateDistribution{
				Key:   "last-user-input",
				Value: event.Timestamp,
			}, event.Building+"-"+event.Room)
		}
	}

	//log.Printf("sent")

	//we need to mark the room to be cheked and updated at the next roomTick
	//roomUpdateChan <- event.Building + "-" + event.Room
}

func distributeHeartbeat(event heartbeat.Event) {
	if event.Category != "Heartbeat" {
		//we don't care
		return
	}

	toSend := StateDistribution{Key: "last-heartbeat", Value: event.Timestamp}

	if runLocal {
		localStateBuffering(toSend, event.Hostname, "device")
		localStateBuffering(StateDistribution{
			Key:   "last-heartbeat-received",
			Value: event.Timestamp,
		}, event.Building+"-"+event.Room, "room")
	} else {
		sendToStateBuffering(toSend, event.Hostname)
	}
}

func SendToStateBuffer(state StateDistribution, hostname string, mapType string) {
	if runLocal {
		localStateBuffering(state, hostname, mapType)
	} else {
		sendToStateBuffering(state, hostname)
	}
}

func localStateBuffering(state StateDistribution, hostname string, mapType string) {

	//check for a nil interface
	if state.Value == nil {
		return
	}

	//check for a blank string
	switch state.Value.(type) {
	case string:
		if len(state.Value.(string)) <= 0 {
			return
		}
	}

	switch mapType {
	case "room":
		bufferLocally(state, hostname, localRoomStateMap)
	case "device":
		bufferLocally(state, hostname, localStateMap)
	}
}

func bufferLocally(state StateDistribution, hostname string, mapToUse map[string]map[string]interface{}) {
	if _, ok := mapToUse[hostname]; ok {

		//pardon the switch statements - you can't use the .(type) assertion in an if statement

		//check to make sure that Value is a Map
		switch state.Value.(type) {
		case map[string]interface{}:
			break
		default:
			mapToUse[hostname][state.Key] = state.Value
			return
		}

		//make sure it even exsists
		if _, ok := mapToUse[hostname][state.Key]; !ok {
			mapToUse[hostname][state.Key] = state.Value
			return
		}

		//make sure we need to do a replace, if there's a type mismatch, we just overwrite
		switch mapToUse[hostname][state.Key].(type) {
		case map[string]interface{}:
			//if val is also a map
			break
		default:
			mapToUse[hostname][state.Key] = state.Value
			return
		}

		var a map[string]interface{}
		a = state.Value.(map[string]interface{})
		var b map[string]interface{}
		b = mapToUse[hostname][state.Key].(map[string]interface{})
		//now we get to compare the child values
		replaceMapValues(&a, &b)
	}

	color.Set(color.FgGreen)
	log.Printf("Adding state map for %v", hostname)
	color.Unset()

	mapToUse[hostname] = make(map[string]interface{})
	mapToUse[hostname][state.Key] = state.Value
}

//here's where we decide if we want to distribute to the child processes or if we want to just put it in a map here
func sendToStateBuffering(state StateDistribution, hostname string) {
	color.Set(color.FgGreen)
	log.Printf("[distributor] Sending to buffer.")
	color.Unset()
	//check if it's in the map
	if val, ok := stateCacheMap[hostname]; ok {
		val <- state
		return
	}
	//we need to add it to the map

	cacheChan := make(chan StateDistribution, 100)
	stateCacheMap[hostname] = cacheChan
	cacheChan <- state

	//now we need to start a aggregator to handle the caching
	//go startAggregator(cacheChan, hostname)
}

//will copy map a to b, adding values and overwriting values as found
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
