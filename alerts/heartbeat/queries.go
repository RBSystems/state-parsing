package heartbeat

const HeartbeatLostQuery = `{
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
  "size": 100
}
`

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
        },
        {
          "match": {
            "alerting": true
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
