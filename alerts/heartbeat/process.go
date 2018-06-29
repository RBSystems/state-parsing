package heartbeat

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/byuoitav/state-parsing/alerts/base"
	"github.com/byuoitav/state-parsing/alerts/device"
	"github.com/byuoitav/state-parsing/alerts/room"
	"github.com/fatih/color"
)

/*
const (
	RESTORED = "heartbeat-restored"
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

*/
/*
	We've got some heartbeats that are lost - we need to verify that they're not suppressing alerts, for themselves or for the room in general.

	Regardless we need to validate that they (and their associated rooms) are marked as alerting.
*/

/*
	for i := range resp.Hits.Hits {
		curHit := resp.Hits.Hits[i].Source

		//add the room to be checked
		roomsToCheck[curHit.Room] = true

		//make sure that it's marked as alerting
		if curHit.Alerting == false || curHit.Alerts[base.LOST_HEARTBEAT].Alerting == false {

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
*/
/*
	Now we need to:
	1) check to see if the rooms in question are suppressing alerts/alerting
	2) update the device/rooms that weren't alerting already to be alerting
*/
/*
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
		if alerting[k] {
			//add it to the list to mark as alerting
			roomsToMark = append(roomsToMark, k)
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
*/

func processHeartbeatRestoredResponse(resp device.HeartbeatRestoredQueryResponse) (map[string][]base.Alert, error) {
	roomsToCheck := make(map[string]bool)
	// devicesToUpdate :=
	deviceIDsToUpdate := []string{}
	alertsByRoom := make(map[string][]base.Alert)
	toReturn := map[string][]base.Alert{}

	// there are no devices that have heartbeats restored
	if len(resp.Hits.Hits) <= 0 {
		log.Printf(color.HiGreenString("[%s] No heartbeats restored"), RESTORED)
		return toReturn, nil
	}

	// loop through all the devices that have had restored heartbeats
	// and create an alert for them
	for i := range resp.Hits.Hits {
		device := resp.Hits.Hits[i].Source

		// get building/room off of hostname
		split := strings.Split(device.Hostname, "-")
		if len(split) != 3 {
			logger.Error("%s is an improper hostname. skipping it...", device.Hostname)
			continue
		}
		building := split[0]
		room := split[1]
		roomKey := building + "-" + room

		// make sure to check this room later
		roomsToCheck[roomKey] = true

		// if it's alerting, we need to set alerting to false
		deviceIDsToUpdate = append(deviceIDsToUpdate, resp.Hits.Hits[i].ID)

		// if a device's alerts aren't suppressed, create the alert
		if !device.Suppress {
			logger.Info("Creating hearbeat restored alert for %s", device.Hostname)
			content, err := json.Marshal(base.SlackAlert{
				Markdown: false,
				Attachments: []base.SlackAttachment{base.SlackAttachment{
					Fallback: fmt.Sprintf("Restored Heartbeat. Device %v sent heartbeat at %v.", device.Hostname, device.LastHeartbeat),
					Title:    "Restored Heartbeat",
					Fields: []base.SlackAlertField{base.SlackAlertField{
						Title: "Device",
						Value: device.Hostname,
						Short: true,
					},
						base.SlackAlertField{
							Title: "Received at",
							Value: device.LastHeartbeat,
							Short: true,
						}},
					Color: "good",
				}}})

			if err != nil {
				log.Printf(color.HiRedString("Couldn't marshal the slack alert for %v. Error: %v", device.Hostname, err.Error()))
				continue
			}

			alert := base.Alert{
				AlertType: base.SLACK,
				Content:   content,
				Device:    device.Hostname,
			}

			if _, ok := alertsByRoom[roomKey]; ok {
				alertsByRoom[roomKey] = append(alertsByRoom[roomKey], alert)
			} else {
				alertsByRoom[roomKey] = []base.Alert{alert}
			}
		} else {
			logger.Warn("Not creating alert for %s, because they are suppressed", device.Hostname)
		}
	}

	// mark devices as not alerting
	device.MarkDevicesAsNotAlerting(deviceIDsToUpdate)

	/* send alerts */
	// get the rooms
	rooms, err := room.GetRoomsBulk(func(vals map[string]bool) []string {
		ret := []string{}
		for k, _ := range vals {
			ret = append(ret, k)
		}
		return ret
	}(roomsToCheck))
	if err != nil {
		logger.Error("error getting rooms: %v", err.Error())
		return toReturn, err
	}

	// figure out if a room's alerts are suppressed
	_, suppressed := AlertingSuppressedRooms(rooms)

	// send alerts to rooms that aren't suppressed
	for room, alerts := range alertsByRoom {
		if !suppressed[room] {
			for _, a := range alerts {
				if _, ok := toReturn[a.AlertType]; ok {
					toReturn[a.AlertType] = append(toReturn[a.AlertType], a)
				} else {
					toReturn[a.AlertType] = []base.Alert{a}
				}
			}
		}
	}

	for k, v := range toReturn {
		logger.Info("%v %v alerts to be sent", len(v), k)
	}

	return toReturn, nil
}
