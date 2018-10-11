package cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/state-parser/elk"
	"github.com/byuoitav/common/state/statedefinition"
)

const maxSize = 10000

const defaultDevIndex = "oit-static-av-devices-v2"

//InitializeCaches initializes the caches with data from ELK
func InitializeCaches() {
	Caches = make(map[string]Cache)

	//get DEFAULT devices
	defaultDevs, err := GetStaticDevices(defaultDevIndex)
	if err != nil {
		log.L.Errorf(err.Addf("Couldn't get information for default device cache").Error())
	} /*
		//get DEFAULT rooms
		defaultRooms, err := GetStaticRooms(defRoomIndx)
		if err != nil {
			log.L.Errorf(err.Addf("Couldn't get information for default room cache").Error())
		}
	*/

	//defaultDevs := []statedefinition.StaticDevice{}
	defaultRooms := []statedefinition.StaticRoom{}

	cache := makeCache(defaultDevs, defaultRooms)
	Caches[DEFAULT] = cache
	log.L.Infof("Default Caches initialized. %v devices and %v rooms", len(defaultDevs), len(defaultRooms))

	/*
		//get DMPS devices
		dmpsDevs, err := GetStaticDevices(dmpsDevIndx)
		if err != nil {
			log.L.Errorf(err.Addf("Couldn't get information for dmps device cache").Error())
		}

		//get DMPS rooms
		dmpsRooms, err := GetStaticRooms(dmpsRoomIndx)
		if err != nil {
			log.L.Errorf(err.Addf("Couldn't get information for dmps room cache").Error())
		}

		cache = makeCache(dmpsDevs, dmpsRooms)
		Caches[DMPS] = &cache

		log.L.Infof("DMPS Caches initialized. %v devices and %v rooms", len(dmpsDevs), len(dmpsRooms))
	*/
}

//GetStaticDevices queries the provided index in ELK and unmarshals the records into a list of static devices
func GetStaticDevices(index string) ([]statedefinition.StaticDevice, *nerr.E) {
	log.L.Debugf("Getting device information from %v", index)
	query := elk.GenericQuery{
		Size: maxSize,
	}

	b, er := json.Marshal(query)
	if er != nil {
		return []statedefinition.StaticDevice{}, nerr.Translate(er).Addf("Couldn't marshal generic query %v", query)
	}

	resp, err := elk.MakeELKRequest("GET", fmt.Sprintf("/%v/_search", index), b)
	if err != nil {
		return []statedefinition.StaticDevice{}, err.Addf("Couldn't retrieve static index %v for cache", index)
	}
	ioutil.WriteFile("/tmp/test", resp, 0777)

	var queryResp elk.StaticDeviceQueryResponse

	er = json.Unmarshal(resp, &queryResp)
	if er != nil {
		return []statedefinition.StaticDevice{}, nerr.Translate(er).Addf("Couldn't unmarshal response from static index %v.", index)
	}

	var toReturn []statedefinition.StaticDevice
	for i := range queryResp.Hits.Wrappers {
		toReturn = append(toReturn, queryResp.Hits.Wrappers[i].Device)
	}

	return toReturn, nil
}

func GetStaticRooms(index string) ([]statedefinition.StaticRoom, *nerr.E) {
	query := elk.GenericQuery{
		Size: maxSize,
	}

	b, er := json.Marshal(query)
	if er != nil {
		return []statedefinition.StaticRoom{}, nerr.Translate(er).Addf("Couldn't marshal generic query %v", query)
	}

	resp, err := elk.MakeELKRequest("GET", fmt.Sprintf("/%v/_search", index), b)
	if err != nil {
		return []statedefinition.StaticRoom{}, err.Addf("Couldn't retrieve static index %v for cache", index)
	}
	log.L.Infof("Getting the info for %v", index)

	var queryResp elk.StaticRoomQueryResponse

	er = json.Unmarshal(resp, &queryResp)
	if er != nil {
		return []statedefinition.StaticRoom{}, nerr.Translate(er).Addf("Couldn't unmarshal response from static index %v.", index)
	}

	var toReturn []statedefinition.StaticRoom
	for i := range queryResp.Hits.Wrappers {
		toReturn = append(toReturn, queryResp.Hits.Wrappers[i].Room)
	}

	return toReturn, nil

}

func makeCache(devices []statedefinition.StaticDevice, rooms []statedefinition.StaticRoom) Cache {

	toReturn := memorycache{}

	//toMerge := []statedefinition.StaticDevice{}

	//go through and create our maps
	toReturn.deviceCache = make(map[string]DeviceItemManager)
	for i := range devices {
		//check for duplicate
		v, ok := toReturn.deviceCache[devices[i].DeviceID]
		if ok {
			continue
		}

		if len(devices[i].DeviceID) < 1 {
			log.L.Errorf("DeviceID cannot be blank. Device: %+v", devices[i])
			continue
		}

		v, err := GetNewDeviceManagerWithDevice(devices[i])
		if err != nil {
			log.L.Errorf("Cannot create device manager for %v: %v", devices[i].DeviceID, err.Error())
			continue
		}

		/*
			if devices[i].DeviceType == "" {
				//get the device type
				a := statedefinition.StaticDevice{
					DeviceID:   devices[i].DeviceID,
					DeviceType: GetDeviceTypeByID(devices[i].DeviceID),
				}
				toMerge = append(toMerge, a)
			}
		*/

		respChan := make(chan DeviceTransactionResponse, 1)
		v.WriteRequests <- DeviceTransactionRequest{
			MergeDeviceEdit: true,
			MergeDevice:     devices[i],
			ResponseChan:    respChan,
		}
		val := <-respChan

		if val.Error != nil {
			log.L.Errorf("Error initializing cache for %v: %v.", devices[i].DeviceID, val.Error.Error())
		}
		toReturn.deviceCache[devices[i].DeviceID] = v
	}
	/*
		//do it at some point
		go func(v []statedefinition.StaticDevice) {
			log.L.Info(color.HiBlueString("Adding device type for %v devices in 10 seconds", len(v)))
			time.Sleep(10 * time.Second)
			log.L.Info(color.HiBlueString("Adding device type for %v devices now", len(v)))
			for _, i := range v {
				dev, _ := GetCache(DEFAULT).GetDeviceRecord(i.DeviceID)
				dev.DeviceType = i.DeviceType
				dev.UpdateTimes["device-type"] = time.Now()

				GetCache(DEFAULT).CheckAndStoreDevice(dev)
			}
		}(toMerge)
	*/

	toReturn.roomCache = make(map[string]RoomItemManager)
	for i := range rooms {
		//check for duplicate
		v, ok := toReturn.roomCache[devices[i].DeviceID]
		if ok {
			continue
		}
		v = GetNewRoomManager(rooms[i].RoomID)

		respChan := make(chan RoomTransactionResponse, 1)
		v.WriteRequests <- RoomTransactionRequest{
			MergeRoom:    rooms[i],
			ResponseChan: respChan,
		}
		val := <-respChan

		if val.Error != nil {
			log.L.Errorf("Error initializing cache for %v: %v.", rooms[i].RoomID, val.Error.Error())
		}
		toReturn.roomCache[rooms[i].RoomID] = v
	}

	return &toReturn
}
