package forwarding

import (
	"fmt"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/state-parser/config"
	"github.com/byuoitav/state-parser/state/forwarding/managers"
)

//BufferManager is meant to handle buffering events/updates to the eventual forever home of the information
type BufferManager interface {
	Send(toSend interface{}) error
}

//Key is made up of the CacheType-DataType-EventType
//e.g. default-device-all or legacy-event-all
var managerMap map[string][]BufferManager

func init() {
	log.L.Infof("Initializing buffer managers")
	c := config.GetConfig()
	managerMap = make(map[string][]BufferManager)
	for _, i := range c.Forwarders {
		curName := fmt.Sprintf(fmt.Sprintf("%v-%v-%v", i.CacheType, i.DataType, i.EventType))
		switch i.Type {
		case config.ELKSTATIC:
			switch i.DataType {
			case config.ROOM:
				managerMap[curName] = append(managerMap[curName], managers.GetDefaultElkStaticRoomForwarder(
					i.Elk.URL,
					GetIndexFunction(i.Elk.IndexPattern, i.Elk.IndexRotationInterval),
					time.Duration(i.Interval)*time.Second,
					i.Elk.Upsert,
				))
			case config.DEVICE:
				managerMap[curName] = append(managerMap[curName], managers.GetDefaultElkStaticDeviceForwarder(
					i.Elk.URL,
					GetIndexFunction(i.Elk.IndexPattern, i.Elk.IndexRotationInterval),
					time.Duration(i.Interval)*time.Second,
					i.Elk.Upsert,
				))
			}
		case config.ELKTIMESERIES:
			managerMap[curName] = append(managerMap[curName], managers.GetDefaultElkTimeSeries(
				i.Elk.URL,
				GetIndexFunction(i.Elk.IndexPattern, i.Elk.IndexRotationInterval),
				time.Duration(i.Interval)*time.Second,
			))
		case config.COUCH:
			managerMap[curName] = append(managerMap[curName], managers.GetDefaultCouchDeviceBuffer(
				i.Couch.URL,
				i.Couch.DatabaseName,
				time.Duration(i.Interval)*time.Second,
			))
		}

	}

	log.L.Infof("Buffer managers initialized")
}

//GetManagersForType a
func GetManagersForType(cacheType, dataType, eventType string) []BufferManager {
	log.L.Debugf("Getting %s managers for %v-%v", cacheType, dataType, eventType)
	v, ok := managerMap[fmt.Sprintf("%s-%s-%s", cacheType, dataType, eventType)]
	if !ok {
		log.L.Errorf("Unknown cache type: %v", cacheType)
		return []BufferManager{}
	}
	return v
}

//GetIndexFunction .
func GetIndexFunction(indexPattern, rotationInterval string) func() string {
	switch rotationInterval {

	case config.DAILY:
		return func() string {
			return fmt.Sprintf("%v-%v", indexPattern, time.Now().Format("20060102"))
		}
	case config.WEEKLY:
		return func() string {
			yr, wk := time.Now().ISOWeek()
			return fmt.Sprintf("%v-%v%v", indexPattern, yr, wk)
		}
	case config.MONTHLY:
		return func() string {
			return fmt.Sprintf("%v-%v", indexPattern, time.Now().Format("200601"))
		}
	case config.YEARLY:
		return func() string {
			return fmt.Sprintf("%v-%v", indexPattern, time.Now().Format("2006"))
		}
	case config.NOROTATE:
		return func() string {
			return indexPattern

		}
	default:
		log.L.Fatalf("Unknown interval %v for index %v", rotationInterval, indexPattern)
	}
	return func() string {
		return indexPattern
	}
}
