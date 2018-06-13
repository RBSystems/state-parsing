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
