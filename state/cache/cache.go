package cache

import (
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	sd "github.com/byuoitav/common/state/statedefinition"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/state-parser/state/forwarding"
)

//Cache is our state cache - it's meant to be a representation of the static indexes
type Cache interface {
	CheckAndStoreDevice(device sd.StaticDevice) (bool, sd.StaticDevice, *nerr.E)
	GetDeviceRecord(deviceID string) (sd.StaticDevice, *nerr.E)
	CheckAndStoreRoom(room sd.StaticRoom) (bool, sd.StaticRoom, *nerr.E)
	GetRoomRecord(roomID string) (sd.StaticRoom, *nerr.E)

	StoreDeviceEvent(toSave sd.State) (bool, sd.StaticDevice, *nerr.E)
	StoreAndForwardEvent(event events.Event) (bool, *nerr.E)
}

var Caches map[string]Cache

func init() {
	log.L.Infof("Initializing Caches")
	//start
	InitializeCaches()
	log.L.Infof("Caches Initialized.")
}

func GetCache(cacheType string) Cache {
	return Caches[cacheType]
}

type memorycache struct {
	deviceCache map[string]DeviceItemManager
	roomCache   map[string]RoomItemManager

	cacheType string
}

func (c *memorycache) StoreAndForwardEvent(v events.Event) (bool, *nerr.E) {
	log.L.Debugf("Event: %+v", v)

	//Forward All
	list := forwarding.GetManagersForType(c.cacheType, forwarding.EVENTALL)
	for i := range list {
		list[i].Send(v)
	}

	//if it's an error, we don't want to try and store it, as it probably won't correlate to a device field
	if HasTag(events.Error, v.EventTags) {
		return false, nil
	}

	//Cache
	changes, newDev, err := c.StoreDeviceEvent(sd.State{
		ID:    v.TargetDevice.DeviceID,
		Key:   v.Key,
		Time:  v.Timestamp,
		Value: v.Value,
		Tags:  v.EventTags,
	})

	if err != nil {
		return false, err.Addf("Couldn't store and forward device event")
	}

	list = forwarding.GetManagersForType(c.cacheType, forwarding.DEVICEALL)
	for i := range list {
		list[i].Send(newDev)
	}

	//if there are changes and it's not a heartbeat event
	if changes && !events.HasTag(v, events.Heartbeat) {

		log.L.Debugf("Event resulted in changes")

		//get the event stuff to forward
		list = forwarding.GetManagersForType(c.cacheType, forwarding.EVENTDELTA)
		for i := range list {
			list[i].Send(v)
		}
		list = forwarding.GetManagersForType(c.cacheType, forwarding.DEVICEDELTA)
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
		log.L.Debugf("Creating a new device manager for %v", toSave.ID)

		//we need to create a new manager and set it up
		manager = GetNewDeviceManager(toSave.ID)
		c.deviceCache[toSave.ID] = manager
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
		return false, sd.StaticDevice{}, resp.Error.Addf("Couldn't store event %v.", toSave)
	}

	return resp.Changes, resp.NewDevice, nil
}

/*CheckAndStoreDevice takes a device, will check to see if there are deltas compared to the values in the map, and store any changes.

Bool returned denotes if there were any changes. True indicates that there were updates
*/
func (c *memorycache) CheckAndStoreDevice(device sd.StaticDevice) (bool, sd.StaticDevice, *nerr.E) {
	if len(device.DeviceID) == 0 {
		return false, sd.StaticDevice{}, nerr.Create("Static Device must have an ID field to be loaded into the databaset", "invalid-device")
	}

	manager, ok := c.deviceCache[device.DeviceID]

	if !ok {
		manager = GetNewDeviceManager(device.DeviceID)
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
		return false, sd.StaticDevice{}, resp.Error.Addf("Couldn't store device %v.", device)
	}

	if resp.Changes {
		list := forwarding.GetManagersForType(c.cacheType, forwarding.DEVICEDELTA)
		for i := range list {
			list[i].Send(resp.NewDevice)
		}

		list = forwarding.GetManagersForType(c.cacheType, forwarding.DEVICEALL)
		for i := range list {
			list[i].Send(resp.NewDevice)
		}
	}

	return resp.Changes, resp.NewDevice, nil
}

//GetDeviceRecord returns a device with the corresponding ID, if any is found in the memorycache
func (c *memorycache) GetDeviceRecord(deviceID string) (sd.StaticDevice, *nerr.E) {

	manager, ok := c.deviceCache[deviceID]
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
	if len(room.RoomID) == 0 {
		return false, sd.StaticRoom{}, nerr.Create("Static room must have a roomID to be compared and stored", "invalid-room")
	}

	manager, ok := c.roomCache[room.RoomID]
	if !ok {
		manager = GetNewRoomManager(room.RoomID)
	}

	respChan := make(chan RoomTransactionResponse, 1)

	//send a request to update
	manager.WriteRequests <- RoomTransactionRequest{
		MergeRoom:    room,
		ResponseChan: respChan,
	}

	//wait for a response
	resp := <-respChan

	if resp.Error != nil {
		return false, sd.StaticRoom{}, resp.Error.Addf("Couldn't store room %v.", room)
	}

	return resp.Changes, resp.NewRoom, nil
}

//GetRoomRecord returns a room
func (c *memorycache) GetRoomRecord(roomID string) (sd.StaticRoom, *nerr.E) {
	manager, ok := c.roomCache[roomID]
	if !ok {
		return sd.StaticRoom{}, nil
	}

	respChan := make(chan sd.StaticRoom, 1)

	manager.ReadRequests <- respChan
	return <-respChan, nil
}
