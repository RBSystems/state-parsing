package update

import (
	"github.com/byuoitav/state-parsing/update/base"
	"github.com/byuoitav/state-parsing/update/room"
	"github.com/byuoitav/state-parsing/update/updater"
)

var updaters = map[string]updater.Interface{}

func init() {
	updaters = make(map[string]updater.Interface)

	// add updaters here
	updaters[base.ROOM_UPDATE] = updater.NewUpdater(&room.RoomUpdater{})
}

func GetUpdater(name string) (updater.Interface, bool) {
	updater, ok := updaters[name]
	return updater, ok
}
