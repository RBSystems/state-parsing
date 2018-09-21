package marking

import (
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/state-parser/state/cache"
	sd "github.com/byuoitav/state-parser/state/statedefinition"
)

func MarkRoomGeneralAlerting(toMark []string, alerting bool) {

	room := sd.StaticRoom{
		UpdateTimes: make(map[string]time.Time),
	}
	room.Alerting = &alerting

	//ship it off to go with the rest
	for i := range toMark {
		room.RoomID = toMark[i]
		_, _, err := cache.GetCache(cache.DEFAULT).CheckAndStoreRoom(room)
		if err != nil {
			log.L.Errorf("Couldn't clear general alert for %v: %v", toMark[i], err.Error())

		}
	}
}
