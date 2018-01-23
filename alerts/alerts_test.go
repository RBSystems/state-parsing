package alerts

import (
	"log"
	"testing"

	"github.com/byuoitav/state-parsing/alerts/base"
)

func TestStuff(t *testing.T) {
	//ha := heartbeat.HeartbeatAlertFactory{}
	//ha.Run(1)

	res, err := base.GetRoomsBulk([]string{"ITB-1101", "CTB-410"})
	if err != nil {
		log.Printf("Error: %v", err.Error())
	}

	log.Printf("%v", len(res))

}
