package cache

import (
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/byuoitav/state-parser/state/forwarding"
	sd "github.com/byuoitav/state-parser/state/statedefinition"
)

//Cache is our state cache - it's meant to be a representation of the static indexes
type Cache interface {
	CheckAndStoreDevice(device sd.StaticDevice) (bool, sd.StaticDevice, *nerr.E)
	GetDeviceRecord(deviceID string) (sd.StaticDevice, *nerr.E)
	CheckAndStoreRoom(room sd.StaticRoom) (bool, sd.StaticRoom, *nerr.E)
	GetRoomRecord(roomID string) (sd.StaticRoom, *nerr.E)

	StoreDeviceEvent(toSave sd.State) (bool, *nerr.E)
	StoreAndForwardDeviceEvent(event elkreporting.ElkEvent) (bool, *nerr.E)
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
	deviceCache map[string]DeviceItemManager

	roomCache map[string]DeviceItemManager
}

func (c *memorycache) StoreAndForwardEvent(v events.Event) (bool, *nerr.E) {

	//Forward All
	list := forwarding.GetManagersForType(forwarding.EVENTALL)
	for i := range list {
		list[i].Send(v)
	}

	//Cache
	changes, newDev, err := StoreDeviceEvent(statedefinition.State{
		ID:    v.TargetDevice.DeviceID,
		Key:   v.Key,
		Time:  v.Timestamp,
		Value: v.Value,
	})

	if err != nil {
		return false, err.Addf("Couldn't store and forward device event")
	}

	//if there are changes
	if changes {
		//get the event stuff to forward
		list = forwarding.GetManagersForType(forwarding.EVENTDELTA)
		for i := range list {
			list[i].Send(v)
		}
		list = forwarding.GetManagersForType(forwarding.DEVICEDELTA)
		for i := range list {
			list[i].Send(newDev)
		}
	}

	return changes, nil
}

/*
	StoreDeviceEvent takes an event (key value) and stores the value in the field defined as key on a device.S
	Defer use to CheckAndStoreDevice for internal use, as there are significant speed gains.
*/
func (c *memorycache) StoreDeviceEvent(toSave sd.State) (bool, sd.StaticDevice, *nerr.E) {

	if len(toSave.ID) < 1 {
		return false, sd.StaticDevice{}, nerr.Create("State must include device ID", "invaid-parameter")
	}

	manager, ok := c.deviceCache[toSave.ID]
	if !ok {
		//we need to create a new manager and set it up
		manager = GetNewManager(toSave.ID)
	}

	respChan := make(chan DeviceTransactionResponse, 1)

	//send a request to update
	manager.WriteRequests <- DeviceTransactionRequest{
		EventEdit:    true,
		Event:        toSave,
		ResponseChan: respChan,
	}

	//wait for a response
	resp := <-respChan

	if resp.Error != nil {
		return false, sd.StaticDevice{}, err.Addf("Couldn't store event %v.", toSave)
	}

	return resp.Changes, resp.NewDevice, nil
}

/*CheckAndStoreDevice takes a device, will check to see if there are deltas compared to the values in the map, and store any changes.

Bool returned denotes if there were any changes. True indicates that there were updates
Device returned contains ONLY the deltas.
*/
func (c *memorycache) CheckAndStoreDevice(device sd.StaticDevice) (bool, sd.StaticDevice, *nerr.E) {
	if len(device.ID) == 0 {
		return false, sd.StaticDevice{}, nerr.Create("Static Device must have an ID field to be loaded into the databaset", "invalid-device")
	}

	manager, ok := c.deviceCache[device.ID]

	if !ok {
		manager = GetNewManager(device.ID)
	}

	respChan := make(chan DeviceTransactionResponse, 1)

	//send a request to update
	manager.WriteRequests <- DeviceTransactionRequest{
		MergeDeviceEdit: true,
		MergeDevice:     device,
		ResponseChan:    respChan,
	}

	//wait for a response
	resp := <-respChan

	if resp.Error != nil {
		return false, sd.StaticDevice{}, err.Addf("Couldn't store event %v.", toSave)
	}

	return resp.Changes, resp.NewDevice, nil
}

//GetDeviceRecord returns a device with the corresponding ID, if any is found in the memorycache
func (c *memorycache) GetDeviceRecord(deviceID string) (sd.StaticDevice, *nerr.E) {

	manager, ok := c.deviceCache[device.ID]
	if !ok {
		return sd.StaticDevice{}, nil
	}

	respChan := make(chan sd.StaticDevice, 1)

	manager.ReadRequests <- respChan
	return <-respChan, nil
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
	return sd.StaticRoom{}, nil
}
