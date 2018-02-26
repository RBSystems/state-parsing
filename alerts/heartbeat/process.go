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

func processHeartbeatLostResponse(resp device.HeartbeatLostQueryResponse) (map[string][]base.Alert, error) {

	roomsToCheck := make(map[string]bool)
	devicesToUpdate := make(map[string]common.DeviceUpdateInfo)
	alertsByRoom := make(map[string][]base.Alert)
	toReturn := map[string][]base.Alert{}

	if len(resp.Hits.Hits) <= 0 {
		log.Printf(color.HiGreenString("[lost-heartbeat] No heartbeats lost"))
		return toReturn, nil
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
		if curHit.Alerting == false || curHit.Alerts[base.LOST_HEARTBEAT].Alerting == false {

			//debug
			//		log.Printf(color.HiYellowString("Need to mark %v as alerting", curHit.Hostname))

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
			Markdown: false,
			Attachments: []base.SlackAttachment{base.SlackAttachment{
				Fallback: fmt.Sprintf("Lost Heartbeat. Device %v stopped sending heartbeats at %v ", curHit.Hostname, curHit.LastHeartbeat),
				Title:    "Lost Heartbeat",
				Fields: []base.SlackAlertField{base.SlackAlertField{
					Title: "Device",
					Value: curHit.Hostname,
					Short: true,
				},
					base.SlackAlertField{
						Title: "Last Heartbeat",
						Value: curHit.LastHeartbeat,
						Short: true,
					}},
				Color: "danger",
			}}})
		if err != nil {
			log.Printf(color.HiRedString("Couldn't marshal the slack alert for %v. Error: %v", curHit.Hostname, err.Error()))
			continue
		}

		//we need to validate before this that the room in question isn't alerting
		toSend := base.Alert{
			AlertType: base.SLACK,
			Content:   content,
			Device:    curHit.Hostname,
		}

		_, ok := alertsByRoom[curHit.Room]
		if ok {
			alertsByRoom[curHit.Room] = append(alertsByRoom[curHit.Room], toSend)
		} else {
			alertsByRoom[curHit.Room] = []base.Alert{toSend}
		}
	}
	/*
		Now we need to:
		1) check to see if the rooms in question are suppressing alerts/alerting
		2) update the device/rooms that weren't alerting already to be alerting
	*/
	rms, err := room.GetRoomsBulk(func(vals map[string]bool) []string {
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

	alerting, suppressed := AlertingSuppressedRooms(rms)

	roomsToMark := []string{}

	//check the rooms that we have in roomsToCheck to validate that we need to mark them as alerting
	for k, _ := range roomsToCheck {
		if _, ok := alerting[k]; !ok {
			//add it to the list to mark as alerting
			roomsToMark = append(roomsToMark, k)
			log.Printf(color.HiBlueString("Need to mark room %v as alerting", k))
		}
	}

	log.Printf(color.HiBlueString("%v rooms to mark as alerting: ", len(roomsToMark)))

	//mark our rooms as alerting
	room.MarkGeneralAlerting(roomsToMark)

	log.Printf(color.HiBlueString("Starting to mark devices as alerting..."))
	for i := range devicesToUpdate {
		//we need to make a copy of the Secondary Alert Structure so we can use it
		secondaryAlertStructure := make(map[string]interface{})
		secondaryAlertStructure["alert-sent"] = time.Now()
		secondaryAlertStructure["alerting"] = true
		secondaryAlertStructure["message"] = fmt.Sprintf("Time elapsed since last heartbeat: %v", devicesToUpdate[i].Info)

		log.Printf(color.HiBlueString("Need to mark device %v as alerting", devicesToUpdate[i]))
		//mark the devices as alerting
		device.MarkAsAlerting([]string{devicesToUpdate[i].Name}, base.LOST_HEARTBEAT, secondaryAlertStructure)
	}

	//now we check to make sure that the alerts we're going to send aren't in rooms that are suppressed - build
	//the list and then return

	for k, v := range alertsByRoom {

		//if the room is suppressing notifications skip these devices
		if v, ok := suppressed[k]; !ok || v {
			continue
		}

		//otherwise add them to the list to be returned
		for i := range v {
			//check if the alert type been included already
			if _, ok := toReturn[v[i].AlertType]; !ok {
				toReturn[v[i].AlertType] = []base.Alert{v[i]}
				continue
			}

			toReturn[v[i].AlertType] = append(toReturn[v[i].AlertType], v[i])
		}
	}

	for k, v := range toReturn {
		log.Printf(color.HiBlueString("%v %v alerts to be sent", len(v), k))
	}

	return toReturn, nil
}

func AlertingSuppressedRooms(toCheck []room.StaticRoom) (map[string]bool, map[string]bool) {

	alerting := make(map[string]bool)
	suppressed := make(map[string]bool)
	//go through each room in the array and check if it's already alerting

	for i := range toCheck {
		alerting[toCheck[i].Room] = toCheck[i].Alerting
		suppressed[toCheck[i].Room] = toCheck[i].Suppressed
	}

	return alerting, suppressed
}
