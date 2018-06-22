package roomupdate

import (
	"github.com/byuoitav/state-parsing/common"
)

var RoomUpdateQuery = &common.ElkQuery{
	Method:   "POST",
	Endpoint: "/oit-static-av-devices,oit-static-av-rooms/_search",
	Query: `{
"_source": false,
  "query": {
    "query_string": {
      "query": "*"
    }
  },
  "aggs": {
    "rooms": {
      "terms": {
        "field": "room",
        "size": 1000
      },
      "aggs": {
        "index": {
          "terms": {
            "field": "_index"
          },
          "aggs": {
            "alerting": {
              "terms": {
                "field": "alerting"
              },
              "aggs": {
                "device-name": {
                  "terms": {
                    "field": "hostname",
                    "size": 100
                  }
                }
              }
            },
            "power": {
              "terms": {
                "field": "power"
              },
              "aggs": {
                "device-name": {
                  "terms": {
                    "field": "hostname",
                    "size": 100
                  }
                }
              }
            }
          }
        }
      }
    }
  },
  "size": 0
}`,
}

type RoomQueryResponse struct {
	Aggregations struct {
		Rooms struct {
			Buckets []struct {
				Bucket

				Index struct {
					Buckets []struct {
						Bucket

						Power struct {
							Buckets []struct {
								Bucket

								DeviceName struct {
									Buckets []struct {
										Bucket
									}
								} `json:"device-name"`
							}
						} `json:"power"`

						Alerting struct {
							Buckets []struct {
								Key int `json:"key"`
								Bucket

								DeviceName struct {
									Buckets []struct {
										Bucket
									}
								} `json:"device-name"`
							}
						} `json:"alerting"`
					}
				} `json:"index"`
			}
		} `json:"rooms"`
	} `json:"aggregations"`
}

type Bucket struct {
	Key      string `json:"key"`
	DocCount int    `json:"doc_count"`
}
