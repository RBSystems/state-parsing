package cache

import (
	"time"

	"github.com/byuoitav/common/nerr"
	sd "github.com/byuoitav/state-parser/state/statedefinition"
)

/*
Device Item Manager handles managing access to a single device in a cache. Changes to the device are submitted via the IncomingWriteChan and reads are submitted via the IncomingReadChan.
*/
type DeviceItemManager struct {
	IncomingWriteChan chan DeviceTransactionRequest //channel to buffer changes to the device.
	IncomingReadChan  chan chan sd.StaticDevice     //channel to buffer requested reads of the device. The current state of the device will be sent down the channel.

	device sd.StaticDevice //the device this manager is managing (imagine that!)
}

//DeviceTransactionRequest is submitted to read/write a the device being managed by this manager
//If both a MergeDevice and an Event are submitted teh MergeDevice will be processed first
type DeviceTransactionRequest struct {
	ResponseChan chan DeviceTransactionResponse
	MergeDevice  sd.StaticDevice // If you want to update the managed device with the values in this device. Note that the lastest edit timestamp field controls which fields will be kept in a merge.
	Event        sd.State        // If you want to store an event and return changes (if any)
}

type DeviceTransactionResponse struct {
	Changes   bool            //if the Transaction Request resulted in changes
	NewDevice sd.StaticDevice //the updated device with the changes included in the Transaction request included
	Error     *nerr.E         //if there were errors
}

func GetNewManager(id string) *DeviceItemManager {
	a := &DeviceItemManager{
		IncomingWriteChan: make(chan DeviceTransactionRequest, 100),
		IncomingReadChan:  make(chan chan sd.StaticDevice, 50),
		device:            sd.StaticDevice{ID: id, UpdateTimes: make(map[string]time.Time)},
	}
	go StartManager(a)
	return a
}

func StartManager(m *DeviceItemManager) {
	for {
		select {
		case write := <-m.IncomingWriteChan:

		case read := <-m.IncomingReadChan:

		}

	}
}
