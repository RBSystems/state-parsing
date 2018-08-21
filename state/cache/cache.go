package cache

import (
	"sync"

	"github.com/byuoitav/common/nerr"
)

//Cache is our state cache - it's meant to be an in-memory representation of the static indexes
type Cache interface {
	CheckAndStoreDevice(device StaticDevice) (bool, StaticDevice, *nerr.E)
	GetDeviceRecord(deviceID string) (StaticDevice, *nerr.E)
	CheckAndStoreRoom(room StaticRoom) (bool, StaticRoom, *nerr.E)
	GetRoomRecord(roomID string) (StaticRoom, *nerr.E)
}

func GetCache() Cache {
	//initialize before returning?
	return &cache{
		deviceCache: make(map[string]SaticDevice),
		roomCache:   make(map[string]StaticRoom),
	}
}

type cache struct {
	deviceLock  sync.RWMuted //lock for the device cache
	deviceCache map[string]StaticDevice

	roomLock  sync.RWMuted //lock for the room cache
	roomCache map[string]StaticRoom
}

/*CheckAndStoreDevice takes a device, will check to see if there are deltas compared to the values in the map, and store any changes.

Bool returned denotes if there were any changes. True indicates that there were updates
Device returned contains ONLY the deltas.
*/
func (c *cache) CheckAndStoreDevice(device StaticDevice) (bool, StaticDevice, *nerr.E) {
	if len(device.ID) == 0 {
		return false, StaticDevice{}, nerr.Create("Static Device must have an ID field to be loaded into the databaset", "invalid-device")
	}

	//get the current value, if any, from the map
	c.deviceLock.RLock()
	v, ok := deviceCache[device.ID]
	c.deviceLock.RUnlock()

	if !ok {
		//we ned to add to the map
	}

	//we need to do a comparison, update any deltas, then return those fields

}

//GetDeviceRecord returns a device with the corresponding ID, if any is found in the cache
func (C *cache) GetDeviceRecord(deviceID string) (StaticDevice, *nerr.E) {

}

/*CheckAndStoreRoom takes a room, will check to see if there are deltas compared to the values in the map, and store any changes.

Bool returned denotes if there were any changes. True indicates that there were updates
Room returned contains ONLY the deltas.
*/
func (c *cache) CheckAndStoreRoom(room StaticRoom) (bool, StaticRoom, *nerr.E) {

}

//GetRoomRecord returns a room
func (C *cache) GetRoomRecord(roomID string) (StaticRoom, *nerr.E) {

}
