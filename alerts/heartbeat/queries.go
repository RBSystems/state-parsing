package heartbeat

const HeartbeatLostQuery = `{
  "query": {
    "bool": {
      "must": [
        {
          "match": {
            "_type": "control-processor" } }
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
          "bool": {
            "must_not": {
              "exists": {
                "field": "alerts.lost-heartbeat.alert-sent"
              }
            }
          }
        }
      ],
      "minimum_should_match": 1,
      "filter": {
        "range": {
          "last-heartbeat": {
            "lte": "now-60s"
          }
        }
      }
    }
  }
}
`
