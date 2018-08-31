package cache

import (
	"sync"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	sd "github.com/byuoitav/state-parser/state/statedefinition"
)

//Cache is our state cache - it's meant to be a representation of the static indexes
type Cache interface {
	CheckAndStoreDevice(device sd.StaticDevice) (bool, sd.StaticDevice, *nerr.E)
	GetDeviceRecord(deviceID string) (sd.StaticDevice, *nerr.E)
	CheckAndStoreRoom(room sd.StaticRoom) (bool, sd.StaticRoom, *nerr.E)
	GetRoomRecord(roomID string) (sd.StaticRoom, *nerr.E)
	StoreDeviceEvent(toSave sd.State) (bool, *nerr.E)
}

const (
	DMPS    = "dmps"
	DEFAULT = "default"
)

var Caches map[string]*Cache

func init() {
	log.L.Infof("Initializing Caches")
	//start
	//	InitializeCaches()
	log.L.Infof("Caches Initialized.")
}

func GetCache(cacheType string) *Cache {
	return Caches[cacheType]
}

type memorycache struct {
	deviceLock  *sync.RWMutex //lock for the device memorycache
	deviceCache map[string]sd.StaticDevice

	roomLock  *sync.RWMutex //lock for the room memorycache
	roomCache map[string]sd.StaticRoom

	DeviceDeltaChannel chan sd.SaticDevice
}

func (c *memorycache) StoreDeviceEvent(toSave sd.State) (bool, *nerr.E) {

	return false, nil
	c.deviceLock.RLock()
	dev := c.deviceCache[toSave.ID]
	c.deviceLock.RUnlock()

	updates, newdev, err := SetDeviceField(toSave.Key, toSave.Value, toSave.Time, dev)
	if err != nil {
		return false, err.Addf("Couldn't store event %v.", toSave)
	}

	//update if necessary
	if updates {
		c.deviceLock.Lock()
		c.deviceCache[newdev.ID] = newdev
		c.deviceLock.Unlock()
	}

	return updates, nil
}

/*CheckAndStoreDevice takes a device, will check to see if there are deltas compared to the values in the map, and store any changes.

Bool returned denotes if there were any changes. True indicates that there were updates
Device returned contains ONLY the deltas.
*/
func (c *memorycache) CheckAndStoreDevice(device sd.StaticDevice) (bool, sd.StaticDevice, *nerr.E) {
	if len(device.ID) == 0 {
		return false, sd.StaticDevice{}, nerr.Create("Static Device must have an ID field to be loaded into the databaset", "invalid-device")
	}

	//get the current value, if any, from the map
	c.deviceLock.RLock()
	v, ok := c.deviceCache[device.ID]
	c.deviceLock.RUnlock()

	if !ok {
		//we ned to add to the map
		if device.UpdateTimes == nil { //initialize update times if necessary
			device.UpdateTimes = make(map[string]time.Time)
		}

		c.deviceLock.Lock()
		c.deviceCache[device.ID] = device
		c.deviceLock.Unlock()

		//return the whole device
		return true, device, nil
	}

	//we need to do a comparison, update any deltas, then return those fields
	diff, merged, changes, err := sd.CompareDevices(v, device)
	if err != nil {
		return false, sd.StaticDevice{}, err.Addf("Couldn't compare devices")
	}
	if !changes {
		return false, diff, nil
	}

	//there were changes to save
	c.deviceLock.Lock()
	c.deviceCache[merged.ID] = device
	c.deviceLock.Unlock()

	return true, diff, nil
}

//GetDeviceRecord returns a device with the corresponding ID, if any is found in the memorycache
func (c *memorycache) GetDeviceRecord(deviceID string) (sd.StaticDevice, *nerr.E) {

	c.deviceLock.RLock()
	v := c.deviceCache[deviceID]
	c.deviceLock.RUnlock()
	if len(v.ID) == 0 {
		return v, nerr.Create("Not found", "not-found")
	}

	return v, nil
}

/*CheckAndStoreRoom takes a room, will check to see if there are deltas compared to the values in the map, and store any changes.

Bool returned denotes if there were any changes. True indicates that there were updates
Room returned contains ONLY the deltas.
*/
func (c *memorycache) CheckAndStoreRoom(room sd.StaticRoom) (bool, sd.StaticRoom, *nerr.E) {
	return false, room, nil
}

//GetRoomRecord returns a room
func (c *memorycache) GetRoomRecord(roomID string) (sd.StaticRoom, *nerr.E) {
	c.roomLock.Lock()
	v := c.roomCache[roomID]
	c.roomLock.RUnlock()
	if len(v.Room) == 0 {
		return v, nerr.Create("Not found", "not-found")
	}
	return v, nil
}

func getIndexesByType(cacheType string) (room, device string) {
	switch cacheType {
	case DEFAULT:
		return "oit-static-av-rooms", "oit-static-av-devices"
	case DMPS:
		return "oit-static-av-rooms-legacy", "oit-static-av-devices-legacy"
	default:
		return "", ""
	}
}
