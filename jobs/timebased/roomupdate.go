package timebased

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/state-parsing/actions/action"
	"github.com/byuoitav/state-parsing/elk"
	"github.com/byuoitav/state-parsing/forwarding"
)

type RoomUpdateJob struct {
}

const (
	ROOM_UPDATE = "room-update"

	roomUpdateQuery = `
	{
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
	}
	`
)

type roomQueryResponse struct {
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

func (r *RoomUpdateJob) Run(context interface{}) []action.Payload {
	log.L.Debugf("Starting room update job...")

	body, err := elk.MakeELKRequest(http.MethodPost, fmt.Sprintf("/%s,%s/_search", elk.DEVICE_INDEX, elk.ROOM_INDEX), []byte(roomUpdateQuery))
	if err != nil {
		log.L.Warn("failed to make elk request to run room update job: %s", err.String())
		return []action.Payload{}
	}

	var data roomQueryResponse
	gerr := json.Unmarshal(body, &data)
	if gerr != nil {
		log.L.Warn("failed to unmarshal elk response to run room update job: %s", gerr)
		return []action.Payload{}
	}

	acts, err := r.processData(data)
	if err != nil {
		log.L.Warnf("failed to process room update data: %s", err.String())
		return acts
	}

	log.L.Debugf("Finished room update job.")
	return acts
}

func (r *RoomUpdateJob) processData(data roomQueryResponse) ([]action.Payload, *nerr.E) {
	log.L.Debugf("[%s] Processing room update data.", ROOM_UPDATE)
	updatePower := make(map[string]string)
	updateAlerting := make(map[string]int)

	for _, room := range data.Aggregations.Rooms.Buckets {
		log.L.Debugf("[%s] Processing room: %s", ROOM_UPDATE, room.Key)

		// make sure both indicies are there
		if len(room.Index.Buckets) > 2 || len(room.Index.Buckets) == 0 {
			indicies := []string{}
			for _, index := range room.Index.Buckets {
				indicies = append(indicies, index.Key)
			}

			log.L.Warnf("[%s] %s has >2 or 0 indicies. ignoring this room. indicies: %s", ROOM_UPDATE, room.Key, indicies)
			continue

		} else if len(room.Index.Buckets) == 1 {
			// one of the indicies is missing
			if room.Index.Buckets[0].Key == elk.DEVICE_INDEX {
				log.L.Infof("%s doesn't have a room index, so I'll create one for it.", room.Key)
			} else if room.Index.Buckets[0].Key == elk.ROOM_INDEX {
				log.L.Warnf("%s doesn't have a device index. this room should probably be deleted.", room.Key)
				continue
			} else {
				log.L.Warnf("%s is missing it's room index and device index. it has index: %v", room.Key, room.Index.Buckets[0].Key)
				continue
			}
		}

		deviceIndex := room.Index.Buckets[0]
		roomIndex := room.Index.Buckets[1]

		poweredOn := false
		alerting := false

		// make sure index's are correct
		if deviceIndex.Key != elk.DEVICE_INDEX {
			deviceIndex = room.Index.Buckets[1]
			roomIndex = room.Index.Buckets[0]
		}

		log.L.Debugf("\tProcessing device index: %v", deviceIndex.Key)

		// check if any devices are powered on
		for _, p := range deviceIndex.Power.Buckets {
			log.L.Debugf("\t\tPower: %v", p.Key)

			if p.Key == elk.POWER_ON {
				poweredOn = true
			}
		}

		// check if any devices are alerting
		for _, a := range deviceIndex.Alerting.Buckets {
			log.L.Debugf("\t\tAlerting: %v", a.Key)

			if a.Key == elk.ALERTING_TRUE {
				alerting = true
			}
		}

		log.L.Debugf("\tProcessing room index: %v", roomIndex.Key)

		if len(roomIndex.Power.Buckets) == 1 {
			log.L.Debugf("\t\troom power set to: %v", roomIndex.Power.Buckets[0].Key)

			if roomIndex.Power.Buckets[0].Key == elk.POWER_STANDBY && poweredOn {
				// the room is in standby, but there is at least one device powered on
				updatePower[room.Key] = elk.POWER_ON
			} else if roomIndex.Power.Buckets[0].Key == elk.POWER_ON && !poweredOn {
				// the room is on, but there are no devices that are powered on
				updatePower[room.Key] = elk.POWER_STANDBY
			}
		} else if len(roomIndex.Power.Buckets) == 0 {
			log.L.Infof("%s doesn't have a power state. i'll create one for it.", room.Key)

			// set the power state to whatever it's supposed to be
			if poweredOn {
				updatePower[room.Key] = elk.POWER_ON
			} else {
				updatePower[room.Key] = elk.POWER_STANDBY
			}
		} else {
			// this room has more than one power state?
			// we'll just skip this room
			log.L.Warnf("room %s has more than one power state. power buckets: %v", room.Key, roomIndex.Power.Buckets)
			continue
		}

		if len(roomIndex.Alerting.Buckets) == 1 {
			log.L.Debugf("\t\troom alerting set to: %v", roomIndex.Alerting.Buckets[0].Key)

			if roomIndex.Alerting.Buckets[0].Key == elk.ALERTING_FALSE && alerting {
				// the room is in not alerting, but there is at least one device alerting
				updateAlerting[room.Key] = elk.ALERTING_TRUE
			} else if roomIndex.Alerting.Buckets[0].Key == elk.ALERTING_TRUE && !alerting {
				// the room is alerting, but there are no devices that are alerting
				updateAlerting[room.Key] = elk.ALERTING_FALSE
			}
		} else if len(roomIndex.Alerting.Buckets) == 0 {
			log.L.Infof("%s doesn't have an alerting state. i'll create one for it.", room.Key)

			// set the power state to whatever it's supposed to be
			if alerting {
				updateAlerting[room.Key] = elk.ALERTING_TRUE
			} else {
				updateAlerting[room.Key] = elk.ALERTING_FALSE
			}
		} else {
			// this room has more than one alerting state?
			// we'll just skip this room
			log.L.Warnf("%s has more than one alerting state. alerting buckets: %v", roomIndex.Key, roomIndex.Power.Buckets)
			continue
		}
	}

	// do stuff with powerOnRooms and alertingRooms

	for room, power := range updatePower {
		// build state
		state := forwarding.StateDistribution{
			Key:   "power",
			Value: power,
		}

		log.L.Infof("marking %s power as %v", room, power)
		forwarding.SendToStateBuffer(state, room, "room")
	}

	for room, alerting := range updateAlerting {
		// build state
		state := forwarding.StateDistribution{
			Key:   "alerting",
			Value: alerting == 1, // to turn it into a bool
		}

		log.L.Infof("marking %s alerting as %v", room, alerting)
		forwarding.SendToStateBuffer(state, room, "room")
	}

	log.L.Debugf("Successfully updated room state.")

	return []action.Payload{}, nil
}
