package managers

import (
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	sd "github.com/byuoitav/state-parser/state/statedefinition"
)

//CouchStaticDevice is just an sd StaticDevice with an _id and a _rev
type CouchStaticDevice struct {
	sd.StaticDevice

	Rev string `json:"_rev"`
	ID  string `json:"_id"`
}

//GetDefaultCouchDeviceBuffer starts and returns a buffer manager
func GetDefaultCouchDeviceBuffer(couchaddr, database string, interval time.Duration) *CouchDeviceBuffer {

	return &CouchDeviceBuffer{}
}

//CouchDeviceBuffer takes a static device and buffers them for storage in couch
type CouchDeviceBuffer struct {
	incomingChannel chan sd.StaticDevice
	buffer          map[string]CouchStaticDevice

	interval  time.Duration
	database  string
	couchaddr string
}

//Send fulfils the manager interface
func (c *CouchDeviceBuffer) Send(toSend interface{}) error {

	dev, ok := toSend.(sd.StaticDevice)
	if !ok {
		return nerr.Create("Invalid type, couch device buffer expects a StaticDevice", "invalid-type")
	}

	c.incomingChannel <- dev

	return nil
}

func (c *CouchDeviceBuffer) start() {

	log.L.Infof("Starting couch buffer for database", c.database)
	ticker := time.NewTicker(c.interval)

	for {
		select {
		case <-ticker.C:
			//send it off
			log.L.Debugf("Sending bulk ELK update for %v", c.database)
			//send

		case dev := <-c.incomingChannel:
			log.L.Debugf("%v", dev)
			//buffer
		}
	}
}

/*
TODO: Make Buffer and send functions, figure out how to get the rev number back for updates, if new, need to create.

Is it worth it to keep the devices here, or just query them every time. Maybe it's dependant on time? Some sort of threshold

*/
