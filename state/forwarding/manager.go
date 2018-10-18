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

	LEGACYEVENTDELTA  = "legacy-elk-delta"
	LEGACYEVENTALL    = "legacy-elk-all"
	LEGACYDEVICEDELTA = "legacy-device-delta"
	LEGACYDEVICEALL   = "legacy-device-all"

	DMPS    = "dmps"
	DEFAULT = "default"
)

var managerMap map[string]map[string][]BufferManager

func init() {
	log.L.Infof("Initializing buffer managers")
	managerMap = make(map[string]map[string][]BufferManager)

	managerMap[DEFAULT] = make(map[string][]BufferManager)
	managerMap[DEFAULT][EVENTDELTA] = getEventDeltaManagers()
	managerMap[DEFAULT][EVENTALL] = getEventAllManagers()
	managerMap[DEFAULT][DEVICEDELTA] = getDeviceDeltaManagers()
	managerMap[DEFAULT][DEVICEALL] = getDeviceAllManagers()

	managerMap[DMPS] = make(map[string][]BufferManager)
	managerMap[DMPS][LEGACYEVENTDELTA] = getLegacyEventDeltaManagers()
	managerMap[DMPS][LEGACYEVENTALL] = getLegacyEventAllManagers()
	managerMap[DMPS][LEGACYDEVICEDELTA] = getLegacyDeviceDeltaManagers()
	managerMap[DMPS][LEGACYDEVICEALL] = getLegacyDeviceAllManagers()
	log.L.Infof("Buffer managers initialized")
}

//GetManagersForType a
func GetManagersForType(cacheType, BufferType string) []BufferManager {
	log.L.Debugf("Getting all managers for %v", cacheType)
	if v, ok := managerMap[cacheType]; ok {
		return v[BufferType]
	}

	log.L.Errorf("Uknown cache type: %v", cacheType)
	return []BufferManager{}
}

/*
	1) Forward to ELK index oit-av-delta-events
*/
func getEventDeltaManagers() []BufferManager {
	//	return []BufferManager{}
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
	//	return []BufferManager{}
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
		managers.GetDefaultCouchDeviceBuffer(
			"https://couchdb-prd.avs.byu.edu",
			"device-state",
			15*time.Second,
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

/*
	1) Forward to ELK index legacy-av-delta-events
*/
func getLegacyEventDeltaManagers() []BufferManager {
	//	return []BufferManager{}
	return []BufferManager{
		//this is the Delta events forwarder
		managers.GetDefaultElkTimeSeries(
			os.Getenv("ELK_DIRECT_ADDRESS"),
			func() string {
				return fmt.Sprintf("legacy-av-delta-events-%v", time.Now().Year())
			},
		//insert other forwarders here
		),
	}
}

/*
	1) Forward to ELK index legacy-av-all-events
*/
func getLegacyEventAllManagers() []BufferManager {
	//	return []BufferManager{}
	return []BufferManager{
		//this is the All events forwarder
		managers.GetDefaultElkTimeSeries(
			os.Getenv("ELK_DIRECT_ADDRESS"),
			func() string {
				return fmt.Sprintf("legacy-av-all-events-%v", time.Now().Format("20060102"))
			},
		//insert other forwarders here
		),
	}
}

/*
	2) Forward to ELK index oit-legacy-static-av-devices-history
	2) Forward to Couch database legacy-device-state
*/
func getLegacyDeviceDeltaManagers() []BufferManager {

	return []BufferManager{
		managers.GetDefaultElkStaticDeviceForwarder(
			os.Getenv("ELK_DIRECT_ADDRESS"),
			func() string {
				return "legacy-oit-static-av-devices-history"
			},
			15*time.Second,
			false,
		),
		managers.GetDefaultCouchDeviceBuffer(
			"https://couchdb-prd.avs.byu.edu",
			"legacy-device-state",
			15*time.Second,
		),
	}
}

/*
	1) Forward to ELK index oit-legacy-static-av-devices-v2
*/
func getLegacyDeviceAllManagers() []BufferManager {

	return []BufferManager{
		//Device static index
		managers.GetDefaultElkStaticDeviceForwarder(
			os.Getenv("ELK_DIRECT_ADDRESS"),
			func() string {
				return "legacy-oit-static-av-devices-v2"
			},
			15*time.Second,
			true,
		),
	}
}
