package forwarding

import (
	"fmt"
	"os"
	"time"

	"github.com/byuoitav/common/log"
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

var managerMap map[string][]BufferManager

func init() {
	log.L.Infof("Initializing buffer managers")
	managerMap = make(map[string][]BufferManager)

	managerMap[EVENTDELTA] = getEventDeltaManagers()
	managerMap[EVENTALL] = getEventAllManagers()
	managerMap[DEVICEDELTA] = getDeviceDeltaManagers()
	log.L.Infof("Buffer managers initialized")
}

func GetManagersForType(Type string) []BufferManager {
	return managerMap[Type]
}

/*
	1) Forward to ELK index oit-av-delta-events
*/
func getEventDeltaManagers() []BufferManager {
	return []BufferManager{
		//this is the Delta events forwarder
		managers.GetDefaultElkTimeSeries(
			os.Getenv("ELK_DIRECT_ADDRESS"),
			func() string {
				return fmt.Sprintf("oit-av-delta-events-%v", time.Now().Year())
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
