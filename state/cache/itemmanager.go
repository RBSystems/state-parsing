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
	WriteRequests chan DeviceTransactionRequest //channel to buffer changes to the device.
	ReadRequests  chan chan sd.StaticDevice
}

//DeviceTransactionRequest is submitted to read/write a the device being managed by this manager
//If both a MergeDevice and an Event are submitted teh MergeDevice will be processed first
type DeviceTransactionRequest struct {
	ResponseChan chan DeviceTransactionResponse

	// If you want to update the managed device with the values in this device. Note that the lastest edit timestamp field controls which fields will be kept in a merge.
	MergeDeviceEdit bool
	MergeDevice     sd.StaticDevice

	// If you want to store an event and return changes (if any)
	EventEdit bool
	Event     sd.State
}

type DeviceTransactionResponse struct {
	Changes   bool            //if the Transaction Request resulted in changes
	NewDevice sd.StaticDevice //the updated device with the changes included in the Transaction request included
	Error     *nerr.E         //if there were errors
}

func GetNewManager(id string) DeviceItemManager {
	a := DeviceItemManager{
		IncomingWriteRequests: make(chan DeviceTransactionRequest, 100),
		IncomingReadRequests:  make(chan chan sd.StaticDevice, 100),
	}

	device := sd.StaticDevice{ID: id, UpdateTimes: make(map[string]time.Time)}
	go StartManager(a, device)
	return a
}

func StartManager(m DeviceItemManager, device sd.StaticDevice) {

	var merged, diff sd.StaticDevie
	var changes bool
	var err *nerr.E

	for {
		select {
		case write := <-m.WriteRequests:

			if write.MergeDeviceEdit {
				if write.MergeDevice.ID != device.ID {
					write.ResponseChan <- DeviceTransactionResponse{Error: nerr.Create("Can't chagne the ID of a device", "invalid-operation"), NewDevice: device, Changes: false}

				}
				diff, merged, changes, err = sd.CompareDevices(device, write.MergeDevice)

				if err != nil {
					write.ResponseChan <- DeviceTransactionResponse{Error: err, Changes: false}
					continue
				}
			}

			if write.EventEdit {
				changes, merged, err = SetDeviceField(
					write.Event.Key,
					write.Event.Value,
					write.Event.Time,
					device,
				)
				if err != nil {
					write.ResponseChan <- DeviceTransactionResponse{Error: err, Changes: false}
					continue
				}

			}

			if changes {
				//only reassign if we have to
				device = merged
			}

			write.ResponseChan <- DeviceTransactionResponse{Error: err, NewDevice: device, Changes: changes}

		case read := <-m.ReadRequests:
			//just send it back
			read <- device
		}
	}
}
