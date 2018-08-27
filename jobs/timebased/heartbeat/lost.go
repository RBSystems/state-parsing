package heartbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/state-parser/actions"
	"github.com/byuoitav/state-parser/actions/action"
	"github.com/byuoitav/state-parser/actions/slack"
	"github.com/byuoitav/state-parser/elk"
	"github.com/byuoitav/state-parser/state/marking"
	"github.com/byuoitav/state-parser/state/statedefinition"
)

type HeartbeatLostJob struct {
}

const (
	elkAlertField  = "lost-heartbeat"
	HEARTBEAT_LOST = "heartbeat-lost"

	heartbeatLostQuery = `{
  "query": {
    "bool": {
      "must": [
        {
          "match": {
            "_type": "control-processor"
          }
        }
      ],
      "should": [
        {
          "range": {
            "alerts.lost-heartbeat.alert-sent": {
              "lte": "now-20m"
            }
          }
        },
        {
          "match": {
            "alerts.lost-heartbeat.alerting": false
          }
        },
        {
          "bool": {
            "must_not": {
              "exists": {
                "field": "alerts.lost-heartbeat.alert-sent"
              }
            }
          }
        }
      ],
      "minimum_should_match": 2,
      "filter": {
        "range": {
          "last-heartbeat": {
            "lte": "now-60s"
          }
        }
      }
    }
  },
  "size": 100}
`
)

type heartbeatLostQueryResponse struct {
	Took     int  `json:"took,omitempty"`
	TimedOut bool `json:"timed_out,omitempty"`
	Shards   struct {
		Total      int `json:"total,omitempty"`
		Successful int `json:"successful,omitempty"`
		Skipped    int `json:"skipped,omitempty"`
		Failed     int `json:"failed,omitempty"`
	} `json:"_shards,omitempty"`
	Hits struct {
		Total    int     `json:"total,omitempty"`
		MaxScore float64 `json:"max_score,omitempty"`
		Hits     []struct {
			Index  string                       `json:"_index,omitempty"`
			Type   string                       `json:"_type,omitempty"`
			ID     string                       `json:"_id,omitempty"`
			Score  float64                      `json:"_score,omitempty"`
			Source statedefinition.StaticDevice `json:"_source,omitempty"`
		} `json:"hits,omitempty"`
	} `json:"hits,omitempty"`
}

func (h *HeartbeatLostJob) Run(context interface{}) []action.Payload {
	log.L.Debugf("Starting heartbeat lost job...")

	body, err := elk.MakeELKRequest(http.MethodPost, fmt.Sprintf("/%s/_search", elk.DEVICE_INDEX), []byte(heartbeatLostQuery))
	if err != nil {
		log.L.Warn("failed to make elk request to run heartbeat lost job: %s", err.String())
		return []action.Payload{}
	}

	var hrresp heartbeatLostQueryResponse
	gerr := json.Unmarshal(body, &hrresp)
	if gerr != nil {
		log.L.Warn("failed to unmarshal elk response to run heartbeat lost job: %s", gerr)
		return []action.Payload{}
	}

	acts, err := h.processResponse(hrresp)
	if err != nil {
		log.L.Warn("failed to process heartbeat lost response: %s", err.String())
		return acts
	}

	log.L.Debugf("Finished heartbeat lost job.")
	return acts
}

func (h *HeartbeatLostJob) processResponse(resp heartbeatLostQueryResponse) ([]action.Payload, *nerr.E) {
	roomsToCheck := make(map[string]bool)
	devicesToUpdate := make(map[string]elk.DeviceUpdateInfo)
	actionsByRoom := make(map[string][]action.Payload)
	toReturn := []action.Payload{}

	if len(resp.Hits.Hits) <= 0 {
		log.L.Infof("[%s] No heartbeats lost", HEARTBEAT_LOST)
		return toReturn, nil
	}

	/*
		We've got some heartbeats that are lost - we need to verify that they're not suppressing alerts, for themselves or for the room in general.
		Regardless, we need to validate that they (and their associated rooms) are marked as alerting.
	*/

	for i := range resp.Hits.Hits {
		curHit := resp.Hits.Hits[i].Source

		//add the room to be checked
		roomsToCheck[curHit.Room] = true

		//make sure that it's marked as alerting
		if !*curHit.Alerting || !curHit.Alerts[elkAlertField].Alerting {
			//we need to mark it to be updated as alerting
			devicesToUpdate[resp.Hits.Hits[i].ID] = elk.DeviceUpdateInfo{
				Name: resp.Hits.Hits[i].ID,
				Info: curHit.LastHeartbeat.Format(time.RFC3339),
			}
		}

		if *curHit.NotificationsSuppressed {
			//we don't actually send the alert
			continue
		}

		//otherwise we create an alert to be returned, for now we just return a slack alert
		slackAttachment := slack.Attachment{
			Fallback: fmt.Sprintf("Lost Heartbeat. Device %v stopped sending heartbeats at %v ", curHit.Hostname, curHit.LastHeartbeat),
			Title:    "Lost Heartbeat",
			Fields: []slack.AlertField{
				slack.AlertField{
					Title: "Device",
					Value: curHit.Hostname,
					Short: true,
				},
				slack.AlertField{
					Title: "Last Heartbeat",
					Value: curHit.LastHeartbeat.Format(time.RFC3339),
					Short: true,
				},
			},
			Color: "danger",
		}

		//we need to validate before this that the room in question isn't alerting
		a := action.Payload{
			Type:    actions.Slack,
			Device:  curHit.Hostname,
			Content: slackAttachment,
		}

		_, ok := actionsByRoom[curHit.Room]
		if ok {
			actionsByRoom[curHit.Room] = append(actionsByRoom[curHit.Room], a)
		} else {
			actionsByRoom[curHit.Room] = []action.Payload{a}
		}
	}

	/*
		Now we need to:
		1) check to see if the rooms in question are suppressing alerts/alerting
		2) update the device/rooms that weren't alerting already to be alerting
	*/
	rms, err := elk.GetRoomsBulk(func(vals map[string]bool) []string {
		toReturn := []string{}
		for k, _ := range vals {
			toReturn = append(toReturn, k)
		}
		return toReturn
	}(roomsToCheck))
	if err != nil {
		return toReturn, err
	}

	alerting, suppressed := elk.AlertingSuppressedRooms(rms)

	roomsToMark := []string{}
	//check the rooms that we have in roomsToCheck to validate that we need to mark them as alerting
	for k, _ := range roomsToCheck {
		if alerting[k] {
			//add it to the list to mark as alerting
			roomsToMark = append(roomsToMark, k)
		}
	}

	// mark rooms as alerting
	log.L.Infof("Marking %v rooms as alerting", len(roomsToMark))
	marking.MarkRoomGeneralAlerting(roomsToMark, true)

	// mark devices as alerting
	log.L.Infof("Marking devices as alerting...")
	for i := range devicesToUpdate {
		//we need to make a copy of the Secondary Alert Structure so we can use it
		secondaryAlertStructure := make(map[string]interface{})
		secondaryAlertStructure["alert-sent"] = time.Now()
		secondaryAlertStructure["alerting"] = true
		secondaryAlertStructure["message"] = fmt.Sprintf("Time elapsed since last heartbeat: %v", devicesToUpdate[i].Info)

		log.L.Debugf("Marking device %v as alerting.", devicesToUpdate[i])
		marking.MarkDevicesAsAlerting([]string{devicesToUpdate[i].Name}, elkAlertField, secondaryAlertStructure)
	}

	//now we check to make sure that the alerts we're going to send aren't in rooms that are suppressed - build
	//the list and then return
	for room, acts := range actionsByRoom {

		// if the room is suppressing notifications skip these devices
		if v, ok := suppressed[room]; !ok || v {
			continue
		}

		//otherwise add them to the list to be returned
		for i := range acts {
			toReturn = append(toReturn, acts[i])
		}
	}

	log.L.Infof("Created %v actions.", len(toReturn))

	return toReturn, nil
}
