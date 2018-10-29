package elk

import (
	"fmt"
	"time"
)

const OIT_AV = "oit-av-events"
const OIT_AV_ALL = "all-oit-av-events"
const OIT_AV_HEARTBEAT = "oit-av-heartbeats"
const DMPS_EVENT = "oit-av-events-legacy"
const DMPS_EVENT_ALL = "all-oit-av-events-legacy"
const DMPS_HEARTBEAT = "oit-av-heartbeats-legacy"

func GenerateIndexName(in string) string {
	switch in {
	case OIT_AV:
		return fmt.Sprintf("%v-%v", OIT_AV, time.Now().Year())
	case OIT_AV_ALL:
		return fmt.Sprintf("%v-%v", OIT_AV_ALL, time.Now().Format("2016-01-02"))
	case OIT_AV_HEARTBEAT:
		year, week := time.Now().ISOWeek()
		return fmt.Sprintf("%v-%v-%v", OIT_AV_HEARTBEAT, year, week)
	case DMPS_HEARTBEAT:
		year, week := time.Now().ISOWeek()
		return fmt.Sprintf("%v-%v-%v", DMPS_HEARTBEAT, year, week)
	case DMPS_EVENT:
		year := time.Now().Year()
		return fmt.Sprintf("%v-%v-%v", DMPS_EVENT, year)
	case DMPS_EVENT_ALL:
		return fmt.Sprintf("%v-%v", DMPS_EVENT_ALL, time.Now().Format("2016-01-02"))
	default:
		return in

	}

}
