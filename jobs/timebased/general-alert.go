package timebased

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/state-parsing/actions/action"
	"github.com/byuoitav/state-parsing/elk"
	"github.com/byuoitav/state-parsing/forwarding"
)

type GeneralAlertClearingJob struct {
}

const (
	GENERAL_ALERT_CLEARING = "general-alert-clearing"

	generalAlertClearingQuery = `{
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
            "alerts.lost-heartbeat.alerting": false
          }
        },
        {
          "match": {
            "alerting": true
          }
        }
      ]
    }
  },
  "size": 100
	}
	`
)

type generalAlertClearingQueryResponse struct {
	Hits struct {
		Total int `json:"total"`
		Hits  []struct {
			Index  string  `json:"_index"`
			Type   string  `json:"_type"`
			ID     string  `json:"_id"`
			Score  float64 `json:"_score"`
			Source struct {
				Alerting              bool   `json:"alerting"`
				Hostname              string `json:"hostname"`
				LastHeartbeat         string `json:"last-heartbeat"`
				EnableNotifications   string `json:"enable-notifications"`
				LastStateRecieved     string `json:"last-state-recieved"`
				SuppressNotifications string `json:"suppress-notifications"`
				Control               string `json:"control"`
				ViewDashboard         string `json:"view-dashboard"`
				Room                  string `json:"room"`
				Alerts                struct {
					LostHeartbeat struct {
						AlertSent string `json:"alert-sent"`
						Message   string `json:"message"`
						Alerting  bool   `json:"alerting"`
					} `json:"lost-heartbeat"`
				} `json:"alerts"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (g *GeneralAlertClearingJob) Run(context interface{}) []action.Action {
	log.L.Debugf("Starting general-alert clearing job")

	//the query is constructed such that only elements that have a general alerting set to true, but no specific alersts return.
	body, err := elk.MakeELKRequest(http.MethodPost, fmt.Sprintf("/%s/_search", elk.DEVICE_INDEX), []byte(roomUpdateQuery))
	if err != nil {
		log.L.Warn("failed to make elk request to run general alert clearing job: %s", err.String())
		return []action.Action{}
	}

	var resp generalAlertClearingQueryResponse
	gerr := json.Unmarshal(body, &resp)
	if err != nil {
		log.L.Warn("couldn't unmarshal elk response to run general alert clearing job: %s", gerr)
		return []action.Action{}
	}

	log.L.Debugf("[%s] Processing response data", GENERAL_ALERT_CLEARING)

	alertcleared := forwarding.StateDistribution{
		Key:   "alerting",
		Value: false,
	}

	//go through and mark each of these rooms as not alerting, in the general
	for _, hit := range resp.Hits.Hits {
		log.L.Debugf("Marking %s as not general not alerting.", hit.ID)
		forwarding.SendToStateBuffer(alertcleared, hit.ID, "device")
	}

	log.L.Debugf("[%s] Finished general alert clearing job.", GENERAL_ALERT_CLEARING)
	return []action.Action{}
}
