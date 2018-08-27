package cache

import (
	"sync"

	"github.com/byuoitav/common/nerr"
)

//Cache is our state cache - it's meant to be a representation of the static indexes
type Cache interface {
	CheckAndStoreDevice(device StaticDevice) (bool, StaticDevice, *nerr.E)
	GetDeviceRecord(deviceID string) (StaticDevice, *nerr.E)
	CheckAndStoreRoom(room StaticRoom) (bool, StaticRoom, *nerr.E)
	GetRoomRecord(roomID string) (StaticRoom, *nerr.E)
}

const (
	DMPS    = "dmps"
	DEFAULT = "default"
)

var Caches map[string]*Cache

func init() {
	//start
	InitializeCaches()
}

func GetCache(cacheType string) Cache {

	return Caches[cacheType]
}

type memorycache struct {
	deviceLock  *sync.RWMuted //lock for the device memorycache
	deviceCache map[string]StaticDevice

	roomLock  *sync.RWMuted //lock for the room memorycache
	roomCache map[string]StaticRoom
}

/*CheckAndStoreDevice takes a device, will check to see if there are deltas compared to the values in the map, and store any changes.

Bool returned denotes if there were any changes. True indicates that there were updates
Device returned contains ONLY the deltas.
*/
func (c *memorycache) CheckAndStoreDevice(device StaticDevice) (bool, StaticDevice, *nerr.E) {
	if len(device.ID) == 0 {
		return false, StaticDevice{}, nerr.Create("Static Device must have an ID field to be loaded into the databaset", "invalid-device")
	}

	//get the current value, if any, from the map
	c.deviceLock.RLock()
	v, ok := deviceCache[device.ID]
	c.deviceLock.RUnlock()

	if !ok {
		//we ned to add to the map
		c.deviceLock.Lock()
		deviceCache[device.ID] = device
		c.deviceLock.Unlock()

		//return the whole device
		return true, device, nil
	}

	//we need to do a comparison, update any deltas, then return those fields
	diff, merged, changes, err := CompareDevices(v, device)
	if err != nil {
		return false, SaticDevice{}, err.Addf("Couldn't compare devices")
	}
	if !changes {
		return false, diff, nil
	}

	//there were changes to save
	c.deviceLock.Lock()
	deviceCache[merged.ID] = device
	c.deviceLock.Unlock()

	return true, diff, nil
}

//GetDeviceRecord returns a device with the corresponding ID, if any is found in the memorycache
func (C *memorycache) GetDeviceRecord(deviceID string) (StaticDevice, *nerr.E) {

	c.deviceLock.RLock()
	v := deviceCache[deviceID]
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
func (c *memorycache) CheckAndStoreRoom(room StaticRoom) (bool, StaticRoom, *nerr.E) {
	return false, room, nil
}

//GetRoomRecord returns a room
func (C *memorycache) GetRoomRecord(roomID string) (StaticRoom, *nerr.E) {
	c.roomLock.Lock
	v := roomCache[roomID]
	c.roomLock.RUnlock()
	if len(v.ID) == 0 {
		return v, nerr.Create("Not found", "not-found")
	}
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
