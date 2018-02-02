package heartbeat

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/byuoitav/state-parsing/alerts/base"
	"github.com/byuoitav/state-parsing/alerts/device"
	"github.com/byuoitav/state-parsing/alerts/room"
	"github.com/byuoitav/state-parsing/common"
	"github.com/fatih/color"
)

func processResponse(resp device.HeartbeatLostQueryResponse) ([]base.Alert, error) {

	roomsToCheck := make(map[string]bool)
	devicesToUpdate := make(map[string]common.DeviceUpdateInfo)
	alertsByRoom := make(map[string][]base.Alert)
	toReturn := []base.Alert{}

	if len(resp.Hits.Hits) <= 0 {
		log.Printf(color.HiGreenString("[Heartbeat-lost] No heartbeats lost"))
		return []base.Alert{}, nil
	}

	/*
		We've got some heartbeats that are lost - we need to verify that they're not suppressing alerts, for themselves or for the room in general.

		Regardless we need to validate that they (and their associated rooms) are marked as alerting.
	*/

	for i := range resp.Hits.Hits {
		curHit := resp.Hits.Hits[i].Source

		//add the room to be checked
		roomsToCheck[curHit.Room] = true

		//make sure that it's marked as alerting
		if curHit.Alerting == false || curHit.Alerts["heartbeat-lost"].Alerting == false {
			//we need to mark it to be updated as alerting
			devicesToUpdate[resp.Hits.Hits[i].ID] = common.DeviceUpdateInfo{
				Name: resp.Hits.Hits[i].ID,
				Info: curHit.LastHeartbeat,
			}
		}

		if curHit.Suppress == true {
			//we don't actually send the alert
			continue
		}

		//otherwise we create an alert to be returned, for now we just return a slack alert
		content, err := json.Marshal(base.SlackAlert{
			Markdown: true,
			Text:     fmt.Sprintf("Heartbeat lost!\nDevice:\t%v\nLastHeartbeat:\t%v", curHit.Hostname, curHit.LastHeartbeat),
		})
		if err != nil {
			log.Printf(color.HiRedString("Couldn't marshal the slack alert for %v. Error: %v", curHit.Hostname, err.Error()))
			continue
		}

		//we need to validate before this that the room in question isn't alerting

		_, ok := alertsByRoom[curHit.Room]
		if ok {
			alertsByRoom[curHit.Room] = append(alertsByRoom[curHit.Room], base.Alert{
				AlertType: base.SLACK,
				Content:   content,
			})
		} else {
			alertsByRoom[curHit.Room] = []base.Alert{base.Alert{
				AlertType: base.SLACK,
				Content:   content,
			}}
		}
	}
	/*
		Now we need to:
		1) check to see if the rooms in question are suppressing alerts/alerting
		2) update the device/rooms that weren't alerting already to be alerting
	*/
	rooms, err := GetRoomsBulk(func(vals map[string]bool) []string {
		toReturn := []string{}
		for k, _ := range vals {
			toReturn = append(toReturn, k)
		}
		return toReturn
	}(roomsToCheck))

	if err != nil {
		log.Printf(color.HiRedString("Error: %v", err.Error()))
		return toReturn, err
	}

	alerting, suppressed := AlertingSuppressedRooms(rooms)

	roomsToMark := []string{}
	//check the rooms that we have in roomsToCheck to validate that we need to mark them as alerting
	for k, _ := range roomsToCheck {
		if _, ok := alerting[k]; !ok {
			//add it to the list to mark as alerting
			roomsToMark = append(roomsToMark, k)
		}
	}

	//mark our rooms as alerting
	rooms.MarkGeneraleAlerting(roomsToMark)

	for i := range devicesToUpdate {
		//we need to make a copy of the Secondary Alert Structure so we can use it
		secondaryAlertStructure := make(map[string]interface{})
		secondaryAlertStructure["alert-sent"] = time.Now()
		secondaryAlertStructure["alerting"] = true
		secondaryAlertStructure["message"] = fmt.Sprintf("Time elapsed since last heartbeat: %v", devicesToUpdate[i].Info)

		//mark the devices as alerting
		device.MarkGeneralAlerting(roomsToMark, base.LOST_HEARTBEAT, secondaryAlertStructure)
	}

	return toReturn, nil
}

func AlertingSuppressedRooms(toCheck []room.StaticRoom, specificAlerts []string) (map[string]bool, map[string]bool) {

	alerting := make(map[string]bool)
	suppressed := make(map[string]bool)
	//go through each room in the array and check if it's already alerting

	for i := range toCheck {
		alerting[toCheck[i].Room] = toCheck[i].Alerting
		suppressed[toCheck[i].Room] = toCheck[i].Suppressed
	}

	return alerting, suppressed
}
