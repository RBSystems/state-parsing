package roomupdate

import (
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
	err := r.getData()
	if err != nil {
	}

	/* process data from ELK */

	return nil
}

func (r *RoomUpdater) getData() error {
	code, body, err := RoomUpdateQuery.MakeELKRequest(r.LogLevel, r.Name)
	r.V("code: %s, body: %s, err: %s", code, body, err)
	return nil
}
