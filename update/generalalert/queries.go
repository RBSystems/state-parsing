package generalalert

import "github.com/byuoitav/state-parsing/common"

var GeneralAlertQuery = &common.ElkQuery{
	Method:   "POST",
	Endpoint: "/oit-static-av-devices/_search",
	Query: `{
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
}`,
}

type GeneralAlertQueryResponse struct {
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
