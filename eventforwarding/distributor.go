package eventforwarding

import (
	"log"

	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/fatih/color"

	heartbeat "github.com/byuoitav/salt-translator-service/elk"
)

//all we really need to distribute is the event info key/value - where the key is the value to update in the index.

type stateDistribution struct {
	Key   string
	Value string
}

//determine if we run distributed or not
const runLocal bool = true

//this is distribution to the outside areas
var stateCacheMap map[string]chan stateDistribution

//this is for local
var localStateMap map[string]map[string]string

//this is for local
var localRoomStateMap map[string]map[string]string

func StartDistributor() {

	log.Printf("[Distributor] Starting")

	//initialize our ingestion channels
	eventIngestionChannel = make(chan elkreporting.ElkEvent, 1000)
	heartbeatIngestionChannel = make(chan heartbeat.Event, 1000)

	stateCacheMap = make(map[string]chan stateDistribution)
	localStateMap = make(map[string]map[string]string)
	localRoomStateMap = make(map[string]map[string]string)

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
			localStateMap = make(map[string]map[string]string)
			localRoomStateMap = make(map[string]map[string]string)
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
	//we need to pull out the values for stateDistributionm
	toSend := stateDistribution{Key: event.Event.Event.EventInfoKey, Value: event.Event.Event.EventInfoValue}

	if runLocal {
		//we need to check if it's a userinput event, if so we need to update the last-user-input field
		localStateBuffering(toSend, event.Building+"-"+event.Room+"-"+event.Event.Event.Device, "device")

		if event.EventCauseString == "USERINPUT" {
			localStateBuffering(stateDistribution{
				Key:   "last-user-input",
				Value: event.Timestamp,
			}, event.Building+"-"+event.Room+"-"+event.Event.Event.Device, "device")

			//we need to update the room as well.
			localStateBuffering(stateDistribution{
				Key:   "last-user-input",
				Value: event.Timestamp,
			}, event.Building+"-"+event.Room, "room")

		}
		localStateBuffering(stateDistribution{
			Key:   "last-state-received",
			Value: event.Timestamp,
		}, event.Building+"-"+event.Room, "room")

		//we need to update the room state

	} else {
		sendToStateBuffering(toSend, event.Building+"-"+event.Room+"-"+event.Event.Event.Device)
		if event.EventCauseString == "USERINPUT" {
			sendToStateBuffering(stateDistribution{
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

	toSend := stateDistribution{Key: "last-heartbeat", Value: event.Data["_stamp"].(string)}

	if runLocal {
		localStateBuffering(toSend, event.Hostname, "device")
		localStateBuffering(stateDistribution{
			Key:   "last-heartbeat-received",
			Value: event.Timestamp,
		}, event.Building+"-"+event.Room, "room")
	} else {
		sendToStateBuffering(toSend, event.Hostname)
	}
}

func localStateBuffering(state stateDistribution, hostname string, mapType string) {
	if len(state.Value) == 0 {
		return
	}

	switch mapType {
	case "room":

		bufferLocally(state, hostname, localRoomStateMap)
	case "device":
		bufferLocally(state, hostname, localStateMap)
	}
}

func bufferLocally(state stateDistribution, hostname string, mapToUse map[string]map[string]string) {
	if val, ok := mapToUse[hostname]; ok {
		val[state.Key] = state.Value
		return
	}
	color.Set(color.FgGreen)
	log.Printf("Adding state map for %v", hostname)
	color.Unset()

	mapToUse[hostname] = make(map[string]string)
	mapToUse[hostname][state.Key] = state.Value
}

//here's where we decide if we want to distribute to the child processes or if we want to just put it in a map here
func sendToStateBuffering(state stateDistribution, hostname string) {
	color.Set(color.FgGreen)
	log.Printf("[distributor] Sending to buffer.")
	color.Unset()
	//check if it's in the map
	if val, ok := stateCacheMap[hostname]; ok {
		val <- state
		return
	}
	//we need to add it to the map

	cacheChan := make(chan stateDistribution, 100)
	stateCacheMap[hostname] = cacheChan
	cacheChan <- state

	//now we need to start a aggregator to handle the caching
	//go startAggregator(cacheChan, hostname)
}
