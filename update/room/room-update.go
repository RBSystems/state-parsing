package room

import (
	"github.com/byuoitav/state-parsing/update/base"
	"github.com/byuoitav/state-parsing/update/updater"
)

type RoomUpdater struct {
	updater.Updater
}

func (r *RoomUpdater) Init() {
	// assign defaults
	r.Name = base.ROOM_UPDATE
	r.LogLevel = 5
}

func (r *RoomUpdater) Run() error {
	/* get data from ELK */

	/* process data from ELK */

	return nil
}

func (r *RoomUpdater) getData() error {
	return nil
}
