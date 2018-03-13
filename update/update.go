package update

import (
	"github.com/byuoitav/state-parsing/update/base"
	"github.com/byuoitav/state-parsing/update/room"
)

var updaters = map[string]Updater{}

type Updater interface {
	Run(loggingLevel int) error
}

func GetUpdater(name string) (Updater, bool) {
	if len(updaters) == 0 {
		updaters = make(map[string]Updater)

		// add the updaters here
		updaters[base.ROOM_UPDATE] = &room.RoomUpdater{}
	}

	updater, ok := updaters[name]
	return updater, ok
}
