package roomupdate

import (
	"encoding/json"

	"github.com/byuoitav/state-parsing/logger"
	"github.com/byuoitav/state-parsing/tasks/names"
	"github.com/byuoitav/state-parsing/update"
)

type RoomUpdater struct {
	update.Updater
}

func (r *RoomUpdater) Init() {
	r.Name = names.ROOM_UPDATE
	r.LogLevel = logger.VERBOSE
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
	for _, room := range data.Aggregations.Rooms.Buckets {
		r.V("processing room: %s", room.Key)

		for _, index := range room.Field["index"]["buckets"] {
			r.V("\tprocessing index: %v", index.Key)
		}
	}
}
