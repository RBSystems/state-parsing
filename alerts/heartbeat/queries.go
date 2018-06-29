package heartbeat

const HeartbeatRestoredQuery = `{  "_source": [
    "hostname",
    "last-heartbeat",
	"notifications-suppressed"], 
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
  }`
