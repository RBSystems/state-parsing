package heartbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/state/statedefinition"
	"github.com/byuoitav/state-parser/actions"
	"github.com/byuoitav/state-parser/actions/action"
	"github.com/byuoitav/state-parser/actions/slack"
	"github.com/byuoitav/state-parser/elk"
	"github.com/byuoitav/state-parser/state/marking"
)

// RestoredJob .
type RestoredJob struct {
}

const (
	HeartbeatRestored = "heartbeat-restored"

	heartbeatRestoredQuery = `
	{
  "_source": [
    "hostname",
    "last-heartbeat",
    "notifications-suppressed"
  ],
  "query": {
    "bool": {
      "must": [
        {
          "match": {
            "_type": "control-processor"
          }
        },
        {
          "match": {
            "alerts.lost-heartbeat.alerting": true
          }
        }
      ],
      "filter": {
        "range": {
          "last-heartbeat": {
            "gte": "now-30s"
          }
        }
      }
    }
  },
  "size": 1000
}
`
)

type heartbeatRestoredQueryResponse struct {
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

// Run runs the job
func (h *RestoredJob) Run(context interface{}, actionWrite chan action.Payload) {
	log.L.Debugf("Starting heartbeat restored job...")

	body, err := elk.MakeELKRequest(http.MethodPost, fmt.Sprintf("/%s/_search", elk.DEVICE_INDEX), []byte(heartbeatRestoredQuery))
	if err != nil {
		log.L.Warn("failed to make elk request to run heartbeat restored job: %s", err.String())
		return
	}

	var hrresp heartbeatRestoredQueryResponse
	gerr := json.Unmarshal(body, &hrresp)
	if gerr != nil {
		log.L.Warn("failed to unmarshal elk response to run heartbeat restored job: %s", gerr)
		return
	}

	err = h.processResponse(hrresp, actionWrite)
	if err != nil {
		log.L.Warn("failed to process heartbeat restored response: %s", err.String())
	}

	log.L.Debugf("Finished heartbeat restored job.")
}

func (h *RestoredJob) processResponse(resp heartbeatRestoredQueryResponse, actionWrite chan action.Payload) *nerr.E {
	roomsToCheck := make(map[string]bool)
	deviceIDsToUpdate := []string{}
	actionsByRoom := make(map[string][]action.Payload)
	//	toReturn := []action.Payload{}

	// there are no devices that have heartbeats restored
	if len(resp.Hits.Hits) <= 0 {
		log.L.Infof("[%s] No heartbeats restored", HeartbeatRestored)
		return nil
	}

	// loop through all the devices that have had restored heartbeats
	// and create an alert for them
	for i := range resp.Hits.Hits {
		device := resp.Hits.Hits[i].Source

		// get building/room off of hostname
		split := strings.Split(device.Hostname, "-")
		if len(split) != 3 {
			log.L.Warnf("%s is an improper hostname. skipping it...", device.Hostname)
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
		if *device.NotificationsSuppressed {
			continue
		}

		slackAttachment := slack.Attachment{
			Fallback: fmt.Sprintf("Restored Heartbeat. Device %v sent heartbeat at %v.", device.Hostname, device.LastHeartbeat),
			Title:    "Restored Heartbeat",
			Fields: []slack.AlertField{
				slack.AlertField{
					Title: "Device",
					Value: device.Hostname,
					Short: true,
				},
				slack.AlertField{
					Title: "Received at",
					Value: device.LastHeartbeat.Format(time.RFC3339),
					Short: true,
				},
			},
			Color: "good",
		}

		a := action.Payload{
			Type:    actions.Slack,
			Device:  device.Hostname,
			Content: slackAttachment,
		}

		if _, ok := actionsByRoom[roomKey]; ok {
			actionsByRoom[roomKey] = append(actionsByRoom[roomKey], a)
		} else {
			actionsByRoom[roomKey] = []action.Payload{a}
		}
	}

	// mark devices as not alerting
	log.L.Infof("Marking %v devices as not alerting", len(deviceIDsToUpdate))
	marking.ClearHeartbeatAlerts(deviceIDsToUpdate)

	/* send alerts */
	// get the rooms
	rooms, err := elk.GetRoomsBulk(func(vals map[string]bool) []string {
		ret := []string{}
		for k := range vals {
			ret = append(ret, k)
		}
		return ret
	}(roomsToCheck))
	if err != nil {
		return err
	}

	// figure out if a room's alerts are suppressed
	_, suppressed := elk.AlertingSuppressedRooms(rooms)

	// send alerts to rooms that aren't suppressed
	for room, acts := range actionsByRoom {

		if v, ok := suppressed[room]; !ok || v {
			continue
		}

		for i := range acts {
			actionWrite <- acts[i]
		}
	}

	return nil
}
