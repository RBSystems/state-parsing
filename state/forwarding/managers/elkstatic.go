package managers

import (
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	sd "github.com/byuoitav/state-parser/state/statedefinition"
)

//Device
type ElkStaticDeviceForwarder struct {
	ElkStaticForwarder
	update          bool
	incomingChannel chan sd.StaticDevice
	buffer          []ElkBulkUpdateItem
}

//room
type ElkStaticRoomForwarder struct {
	ElkStaticForwarder
	update          bool
	incomingChannel chan sd.StaticRoom
	buffer          []ElkBulkUpdateItem
}

//base
type ElkStaticForwarder struct {
	interval time.Duration //how often to send an update
	url      string
	index    func() string //function to get the indexA
}

func GetDefaultElkStaticDeviceForwarder(URL string, index func() string, interval time.Duration, update bool) *ElkStaticDeviceForwarder {
	toReturn := &ElkStaticDeviceForwarder{
		ElkStaticForwarder: ElkStaticForwarder{
			interval: interval,
			url:      URL,
			index:    index,
		},
		update:          update,
		incomingChannel: make(chan sd.StaticDevice, 10000),
		buffer:          []ElkBulkUpdateItem{},
	}

	go toReturn.start()

	return toReturn
}

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

func GetDefaultElkStaticRoomForwarder(URL string, index func() string, interval time.Duration, update bool) *ElkStaticRoomForwarder {
	toReturn := &ElkStaticRoomForwarder{
		ElkStaticForwarder: ElkStaticForwarder{
			interval: interval,
			url:      URL,
			index:    index,
		},
		incomingChannel: make(chan sd.StaticRoom, 10000),
		buffer:          []ElkBulkUpdateItem{},
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

			go forward(e.url, e.buffer)
			e.buffer = []ElkBulkUpdateItem{}

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

			go forward(e.url, e.buffer)
			e.buffer = []ElkBulkUpdateItem{}

		case event := <-e.incomingChannel:
			e.bufferevent(event)
		}
	}
}

func (e *ElkStaticDeviceForwarder) bufferevent(event sd.StaticDevice) {
	Header := HeaderIndex{
		Index: e.index(),
		Type:  "av-device",
	}
	if e.update {
		Header.ID = event.DeviceID
	}
	e.buffer = append(e.buffer, ElkBulkUpdateItem{
		Header: ElkUpdateHeader{Index: Header},
		Doc:    event,
	})
}

func (e *ElkStaticRoomForwarder) bufferevent(event sd.StaticRoom) {
	Header := HeaderIndex{
		Index: e.index(),
		Type:  "av-room",
	}
	if e.update {
		Header.ID = event.RoomID
	}
	e.buffer = append(e.buffer, ElkBulkUpdateItem{
		Header: ElkUpdateHeader{Index: Header},
		Doc:    event,
	})
}
