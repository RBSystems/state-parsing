package forwarding

import (
	"fmt"
	"os"
	"time"

	"github.com/byuoitav/state-parser/state/forwarding/managers"
)

type BufferManager interface {
	Send(toSend interface{}) error
}

const (
	EVENTDELTA  = "elk-delta"
	EVENTALL    = "elk-all"
	DEVICEDELTA = "device-delta"
)

func GetManagersForType(string Type) []BufferManager {
	switch Type {
	case EVENTDELTA:
		return getEventDeltaManagers()
	case EVENTALL:
		return getEventAllManagers()
	case DEVICEDELTA:
	}

	return []BufferManager{}
}

/*
	1) Forward to ELK index oit-av-delta-events
*/
func getEventDeltaManagers() []BufferManager {
	return []BufferManager{
		//this is the Delta events forwarder
		managers.GetDefaultElkTimeSeries(
			os.Getenv("ELK_DIRECT_ADDRESS"),
			func() {
				return fmt.Sprintf("oit-av-delta-events-%v", time.Now().Year)
			},
		//insert other forwarders here
		),
	}
}

/*
	1) Forward to ELK index oit-av-all-events
*/
func getEventAllManagers() []BufferManager {
	return []BufferManager{
		//this is the All events forwarder
		managers.GetDefaultElkTimeSeries(
			os.Getenv("ELK_DIRECT_ADDRESS"),
			func() {
				return fmt.Sprintf("oit-av-all-events-%v", time.Now().Format("20060102"))
			},
		//insert other forwarders here
		),
	}
	return []BufferManager{}
}

/*
	1) Forward to ELK index oit-static-av-devices
	2) Forward to ELK index oit-static-av-devices-history
	2) Forward to Couch database oit-static-av-devices
*/
func getDeviceDeltaManagers() []BufferManager {

	return []BufferManager{}
}
