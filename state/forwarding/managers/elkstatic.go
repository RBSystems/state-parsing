package managers

import (
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	sd "github.com/byuoitav/state-parser/state/statedefinition"
)

//ElkStaticDeviceForwarder is for a device
type ElkStaticDeviceForwarder struct {
	ElkStaticForwarder
	update          bool
	incomingChannel chan sd.StaticDevice
	buffer          map[string]ElkBulkUpdateItem
}

//ElkStaticRoomForwarder is for rooms
type ElkStaticRoomForwarder struct {
	ElkStaticForwarder
	update          bool
	incomingChannel chan sd.StaticRoom
	buffer          map[string]ElkBulkUpdateItem
}

//ElkStaticForwarder is the common stuff
type ElkStaticForwarder struct {
	interval time.Duration //how often to send an update
	url      string
	index    func() string //function to get the indexA
}

//GetDefaultElkStaticDeviceForwarder returns a regular static device forwarder with a buffer size of 10000
func GetDefaultElkStaticDeviceForwarder(URL string, index func() string, interval time.Duration, update bool) *ElkStaticDeviceForwarder {
	toReturn := &ElkStaticDeviceForwarder{
		ElkStaticForwarder: ElkStaticForwarder{
			interval: interval,
			url:      URL,
			index:    index,
		},
		update:          update,
		incomingChannel: make(chan sd.StaticDevice, 10000),
		buffer:          make(map[string]ElkBulkUpdateItem),
	}

	go toReturn.start()

	return toReturn
}

//Send takes a device and adds it to the buffer
func (e *ElkStaticDeviceForwarder) Send(toSend interface{}) error {

	var event sd.StaticDevice

	switch e := toSend.(type) {
	case *sd.StaticDevice:
		event = *e
	case sd.StaticDevice:
		event = e
	default:
		return nerr.Create("Invalid type to send via an Elk device Forwarder, must be a static device as defined in byuoitav/state-parser/state/statedefinition", "invalid-type")
	}

	e.incomingChannel <- event

	return nil
}

//Send takes a room and adds it to the buffer
func (e *ElkStaticRoomForwarder) Send(toSend interface{}) error {

	var event sd.StaticRoom

	switch e := toSend.(type) {
	case *sd.StaticRoom:
		event = *e
	case sd.StaticRoom:
		event = e
	default:
		return nerr.Create("Invalid type to send via an Elk device Forwarder, must be a static device as defined in byuoitav/state-parser/state/statedefinition", "invalid-type")
	}

	e.incomingChannel <- event

	return nil
}

//GetDefaultElkStaticRoomForwarder returns a regular static room forwarder with a buffer size of 10000
func GetDefaultElkStaticRoomForwarder(URL string, index func() string, interval time.Duration, update bool) *ElkStaticRoomForwarder {
	toReturn := &ElkStaticRoomForwarder{
		ElkStaticForwarder: ElkStaticForwarder{
			interval: interval,
			url:      URL,
			index:    index,
		},
		incomingChannel: make(chan sd.StaticRoom, 10000),
		buffer:          make(map[string]ElkBulkUpdateItem),
		update:          update,
	}

	go toReturn.start()

	return toReturn
}

func (e *ElkStaticDeviceForwarder) start() {
	log.L.Infof("Starting device forwarder for %v", e.index())
	ticker := time.NewTicker(e.interval)

	for {
		select {
		case <-ticker.C:
			//send it off
			log.L.Debugf("Sending bulk ELK update for %v", e.index())

			go prepAndForward(e.url, e.buffer)
			e.buffer = make(map[string]ElkBulkUpdateItem)

		case event := <-e.incomingChannel:
			e.bufferevent(event)
		}
	}
}

func (e *ElkStaticRoomForwarder) start() {
	log.L.Infof("Starting room forwarder for %v", e.index())
	ticker := time.NewTicker(e.interval)

	for {
		select {
		case <-ticker.C:
			//send it off
			log.L.Debugf("Sending bulk ELK update for %v", e.index())

			go prepAndForward(e.url, e.buffer)
			e.buffer = make(map[string]ElkBulkUpdateItem)

		case event := <-e.incomingChannel:
			e.bufferevent(event)
		}
	}
}

func (e *ElkStaticDeviceForwarder) bufferevent(event sd.StaticDevice) {

	//check to see if we already have one for this device
	v, ok := e.buffer[event.DeviceID]
	if !ok {
		Header := HeaderIndex{
			Index: e.index(),
			Type:  "av-device",
		}
		if e.update {
			Header.ID = event.DeviceID
		}
		e.buffer[event.DeviceID] = ElkBulkUpdateItem{
			Header: ElkUpdateHeader{Index: Header},
			Doc:    event,
		}
	} else {
		v.Doc = event
		e.buffer[event.DeviceID] = v
	}

}

func (e *ElkStaticRoomForwarder) bufferevent(event sd.StaticRoom) {
	v, ok := e.buffer[event.RoomID]
	if !ok {
		Header := HeaderIndex{
			Index: e.index(),
			Type:  "av-room",
		}
		if e.update {
			Header.ID = event.RoomID
		}
		e.buffer[event.RoomID] = ElkBulkUpdateItem{
			Header: ElkUpdateHeader{Index: Header},
			Doc:    event,
		}
	} else {
		v.Doc = event
		e.buffer[event.RoomID] = v
	}
}

func prepAndForward(url string, vals map[string]ElkBulkUpdateItem) {
	var toUpdate []ElkBulkUpdateItem
	for _, v := range vals {
		toUpdate = append(toUpdate, v)
	}

	forward(url, toUpdate)
}
