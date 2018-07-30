package timebased

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/state-parser/actions/action"
	"github.com/byuoitav/state-parser/elk"
	"github.com/byuoitav/state-parser/forwarding"
)

type GeneralAlertClearingJob struct {
}

const (
	GENERAL_ALERT_CLEARING = "general-alert-clearing"

	generalAlertClearingQuery = `{
	"_source": [
		"hostname"
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
			ID     string `json:"_id"`
			Source struct {
				Hostname string `json:"hostname"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (g *GeneralAlertClearingJob) Run(context interface{}) []action.Payload {
	log.L.Debugf("Starting general-alert clearing job")

	// The query is constructed such that only elements that have a general alerting set to true, but no specific alersts return.
	body, err := elk.MakeELKRequest(http.MethodPost, fmt.Sprintf("/%s/_search", elk.DEVICE_INDEX), []byte(generalAlertClearingQuery))
	if err != nil {
		log.L.Warn("failed to make elk request to run general alert clearing job: %s", err.String())
		return []action.Payload{}
	}

	var resp generalAlertClearingQueryResponse
	gerr := json.Unmarshal(body, &resp)
	if err != nil {
		log.L.Warn("couldn't unmarshal elk response to run general alert clearing job: %s", gerr)
		return []action.Payload{}
	}

	log.L.Debugf("[%s] Processing response data", GENERAL_ALERT_CLEARING)

	alertcleared := forwarding.State{
		Key:   "alerting",
		Value: false,
	}

	// go through and mark each of these rooms as not alerting, in the general
	for _, hit := range resp.Hits.Hits {
		log.L.Debugf("Marking general alerting on %s as false.", hit.ID)
		alertcleared.ID = hit.ID
		forwarding.BufferState(alertcleared, "device")
	}

	log.L.Debugf("[%s] Finished general alert clearing job.", GENERAL_ALERT_CLEARING)
	return []action.Payload{}
}
