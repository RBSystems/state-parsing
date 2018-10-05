package forwarding

import (
	"fmt"
	"os"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/state-parser/state/forwarding/managers"
)

//BufferManager is meant to handle buffering events/updates to the eventual forever home of the information
type BufferManager interface {
	Send(toSend interface{}) error
}

const (
	EVENTDELTA  = "elk-delta"
	EVENTALL    = "elk-all"
	DEVICEDELTA = "device-delta"
	DEVICEALL   = "device-all"
)

var managerMap map[string][]BufferManager

func init() {
	log.L.Infof("Initializing buffer managers")
	managerMap = make(map[string][]BufferManager)

	managerMap[EVENTDELTA] = getEventDeltaManagers()
	managerMap[EVENTALL] = getEventAllManagers()
	managerMap[DEVICEDELTA] = getDeviceDeltaManagers()
	managerMap[DEVICEALL] = getDeviceAllManagers()
	log.L.Infof("Buffer managers initialized")
}

//GetManagersForType a
func GetManagersForType(Type string) []BufferManager {
	log.L.Debugf("Getting all managers for %v", Type)
	return managerMap[Type]
}

/*
	1) Forward to ELK index oit-av-delta-events
*/func getEventDeltaManagers() []BufferManager {
	return []BufferManager{
		//this is the Delta events forwarder
		managers.GetDefaultElkTimeSeries(
			os.Getenv("ELK_DIRECT_ADDRESS"),
			func() string {
				return fmt.Sprintf("av-delta-events-%v", time.Now().Year())
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
			func() string {
				return fmt.Sprintf("av-all-events-%v", time.Now().Format("20060102"))
			},
		//insert other forwarders here
		),
	}
}

/*
	2) Forward to ELK index oit-static-av-devices-history
	2) Forward to Couch database oit-static-av-devices
*/
func getDeviceDeltaManagers() []BufferManager {

	return []BufferManager{
		managers.GetDefaultElkStaticDeviceForwarder(
			os.Getenv("ELK_DIRECT_ADDRESS"),
			func() string {
				return "oit-static-av-devices-history"
			},
			15*time.Second,
			false,
		),
	}
}

/*
	1) Forward to ELK index oit-static-av-devices
*/
func getDeviceAllManagers() []BufferManager {

	return []BufferManager{
		//Device static index
		managers.GetDefaultElkStaticDeviceForwarder(
			os.Getenv("ELK_DIRECT_ADDRESS"),
			func() string {
				return "oit-static-av-devices-v2"
			},
			15*time.Second,
			true,
		),
	}
}
