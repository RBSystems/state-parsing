package roomupdate

import (
	"encoding/json"

	"github.com/byuoitav/state-parsing/eventforwarding"
	"github.com/byuoitav/state-parsing/logger"
	"github.com/byuoitav/state-parsing/tasks/names"
	"github.com/byuoitav/state-parsing/update"
)

const (
	ROOM_INDEX   = "oit-static-av-rooms"
	DEVICE_INDEX = "oit-static-av-devices"

	STANDBY      = "standby"
	ON           = "on"
	ALERTING     = 1
	NOT_ALERTING = 0
)

type RoomUpdater struct {
	update.Updater
}

func (r *RoomUpdater) Init() {
	r.Name = names.ROOM_UPDATE
	r.LogLevel = logger.INFO
}

func (r *RoomUpdater) Run() error {
	/* get data from ELK */
	body, err := RoomUpdateQuery.MakeELKRequest(r.LogLevel, r.Name)
	if err != nil {
		r.E("error with the initial query: %s", err)
		return err
	}

	var data RoomQueryResponse

	err = json.Unmarshal(body, &data)
	if err != nil {
		r.E("couldn't unmarshal response: %s", err)
		return err
	}

	/* process data from ELK */
	r.processData(data)

	return nil
}

func (r *RoomUpdater) processData(data RoomQueryResponse) {
	r.I("Processing room update data.")
	updatePower := make(map[string]string)
	updateAlerting := make(map[string]int)

	for _, room := range data.Aggregations.Rooms.Buckets {
		r.V("Processing room: %s", room.Key)

		// make sure both indicies are there
		if len(room.Index.Buckets) > 2 || len(room.Index.Buckets) == 0 {
			// there are no indicies/>2 incicies

			indicies := []string{}
			for _, index := range room.Index.Buckets {
				indicies = append(indicies, index.Key)
			}

			r.E("%s has more than >2 or 0 incidies. ignoring this room...", room.Key)
			r.E("indicies of %s: %v", room.Key, indicies)
			continue

		} else if len(room.Index.Buckets) == 1 {
			// one of the indicies is missing

			if room.Index.Buckets[0].Key == DEVICE_INDEX {
				r.W("%s doesn't have a room index. i'll create it.", room.Key)
			} else if room.Index.Buckets[0].Key == ROOM_INDEX {
				r.E("%s doesn't have a device index. this room probably should be deleted", room.Key)
				continue
			} else {
				r.E("%s is missing it's room index and device index. it has index: %v", room.Key, room.Index.Buckets[0].Key)
				continue
			}
		}

		deviceIndex := room.Index.Buckets[0]
		roomIndex := room.Index.Buckets[1]

		poweredOn := false
		alerting := false

		// make sure index's are correct
		if deviceIndex.Key != DEVICE_INDEX {
			deviceIndex = room.Index.Buckets[1]
			roomIndex = room.Index.Buckets[0]
		}

		r.V("processing device index: %v", deviceIndex.Key)

		// check if any devices are powered on
		for _, p := range deviceIndex.Power.Buckets {
			r.V("power: %v", p.Key)

			if p.Key == ON {
				poweredOn = true
			}
		}

		// check if any devices are alerting
		for _, a := range deviceIndex.Alerting.Buckets {
			r.V("alerting: %v", a.Key)

			if a.Key == ALERTING {
				alerting = true
			}
		}

		r.V("processing room index: %v", roomIndex.Key)

		if len(roomIndex.Power.Buckets) == 1 {
			r.V("room power set to: %v", roomIndex.Power.Buckets[0].Key)

			if roomIndex.Power.Buckets[0].Key == STANDBY && poweredOn {
				// the room is in standby, but there is at least one device powered on
				updatePower[room.Key] = ON
			} else if roomIndex.Power.Buckets[0].Key == ON && !poweredOn {
				// the room is on, but there are no devices that are powered on
				updatePower[room.Key] = STANDBY
			}
		} else if len(roomIndex.Power.Buckets) == 0 {
			r.W("%s doesn't have a power state. i'll create one for it.", room.Key)

			// set the power state to whatever it's supposed to be
			if poweredOn {
				updatePower[room.Key] = ON
			} else {
				updatePower[room.Key] = STANDBY
			}
		} else {
			// this room has more than one power state?
			// we'll just skip this room
			r.W("room %s has more than one power state. power buckets: %v", room.Key, roomIndex.Power.Buckets)
			continue
		}

		if len(roomIndex.Alerting.Buckets) == 1 {
			r.V("room alerting set to: %v", roomIndex.Alerting.Buckets[0].Key)

			if roomIndex.Alerting.Buckets[0].Key == NOT_ALERTING && alerting {
				// the room is in not alerting, but there is at least one device alerting
				updateAlerting[room.Key] = ALERTING
			} else if roomIndex.Alerting.Buckets[0].Key == ALERTING && !alerting {
				// the room is alerting, but there are no devices that are alerting
				updateAlerting[room.Key] = NOT_ALERTING
			}
		} else if len(roomIndex.Alerting.Buckets) == 0 {
			r.W("%s doesn't have an alerting state. i'll create one for it.", room.Key)

			// set the power state to whatever it's supposed to be
			if alerting {
				updateAlerting[room.Key] = ALERTING
			} else {
				updateAlerting[room.Key] = NOT_ALERTING
			}
		} else {
			// this room has more than one alerting state?
			// we'll just skip this room
			r.W("%s has more than one alerting state. alerting buckets: %v", roomIndex.Key, roomIndex.Power.Buckets)
			continue
		}
	}

	// do stuff with powerOnRooms and alertingRooms

	for room, power := range updatePower {
		// build state
		state := eventforwarding.StateDistribution{
			Key:   "power",
			Value: power,
		}

		r.I("marking %s power as %v", room, power)
		eventforwarding.SendToStateBuffer(state, room, "room")
	}

	for room, alerting := range updateAlerting {
		// build state
		state := eventforwarding.StateDistribution{
			Key:   "alerting",
			Value: alerting == 1, // to turn it into a bool
		}

		r.I("marking %s alerting as %v", room, alerting)
		eventforwarding.SendToStateBuffer(state, room, "room")
	}

	r.I("Successfully updated room state.")
}
