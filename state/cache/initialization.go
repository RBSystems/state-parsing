package cache

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/state-parser/elk"
	"github.com/byuoitav/state-parser/state/statedefinition"
)

const MAX_SIZE = 10000

func InitializeCaches() {
	Caches = make(map[string]*Cache)

	defDevIndx, defRoomIndx := getIndexesByType(DEFAULT)
	dmpsDevIndx, dmpsRoomIndx := getIndexesByType(DMPS)

	//get DEFAULT devices
	defaultDevs, err := Getstatedefinition.StaticDevices(defDevIndx)
	if err != nil {
		log.L.Erorrf(err.Addf("Couldn't get information for default device cache").Error())
	}

	//get DEFAULT rooms
	defaultRooms, err := Getstatedefinition.StaticRoomst(defRoomIndx)
	if err != nil {
		log.L.Erorrf(err.Addf("Couldn't get information for default room cache").Error())
	}

	cache := makeCache(defaultDevs, defaultRooms)
	Caches[DEFAULT] = cache

	//get DMPS devices
	dmpsDevs, err := Getstatedefinition.StaticDevices(dmpsDevIndx)
	if err != nil {
		log.L.Erorrf(err.Addf("Couldn't get information for dmps device cache").Error())
	}

	//get DMPS rooms
	dmpsRooms, err := Getstatedefinition.StaticRoomst(dmpsRoomIndx)
	if err != nil {
		log.L.Erorrf(err.Addf("Couldn't get information for dmps room cache").Error())
	}

	cache = makeCache(dmpsDevs, dmpsRooms)
	Caches[DMPS] = cache
}

func GetStaticDevices(index string) ([]statedefinition.StaticDevice, *nerr.E) {
	query := elk.GenericQuery{
		Size: MAX_SIZE,
	}

	b, er := json.Marshal(query)
	if er != nil {
		return []statedefinition.StaticDevice{}, nerr.Translate(er).Addf("Couldn't marshal generic query %v", query)
	}

	resp, err := MakeELKRequest("GET", fmt.Sprintf("%v%v/_search", elk.APIAddr, index), b)
	if err != nil {
		return []statedefinition.StaticDevice{}, err.Addf("Couldn't retrieve static index %v for cache", index)
	}

	var queryResp elk.StaticDeviceQueryResponse

	er := json.Unmarshal(resp, &queryResp)
	if er != nil {
		return []statedefinition.StaticDevice{}, nerr.Translate(er).Addf("Couldn't unmarshal response from static index %v.", index)
	}

	//take our query resp, and dump back the Devices
	return queryResp.Hits.Wrappers.Devices, nil
}

func GetStaticRooms(index string) ([]statedefinition.StaticRoom, *nerr.E) {
	query := elk.GenericQuery{
		Size: MAX_SIZE,
	}

	b, er := json.Marshal(query)
	if er != nil {
		return []statedefinition.StaticRoom{}, nerr.Translate(er).Addf("Couldn't marshal generic query %v", query)
	}

	resp, err := MakeELKRequest("GET", fmt.Sprintf("%v%v/_search", elk.APIAddr, index), b)
	if err != nil {
		return []statedefinition.StaticRoom{}, err.Addf("Couldn't retrieve static index %v for cache", index)
	}

	var queryResp elk.StaticRoomQueryResponse

	er := json.Unmarshal(resp, &queryResp)
	if er != nil {
		return []statedefinition.StaticRoom{}, nerr.Translate(er).Addf("Couldn't unmarshal response from static index %v.", index)
	}

	//take our query resp, and dump back the Devices
	return queryResp.Hits.Wrappers.Rooms, nil

}

func makeCache(devices []statedefinition.StaticDevice, rooms []statedefinition.StaticRoom) Cache {
	toReturn := memoryCache{
		deviceLock: &sync.RWMutex{},
		roomLock:   &sync.RWMutex{},
	}

	//go through and create our maps
	deviceMap := make(map[string]statedefinition.StaticDevice)
	for i := range devices {
		deviceMap[devices[i].ID] = devices[i]
	}
	toReturn.deviceCache = deviceMap

	roomMap := make(map[string]statedefinition.StaticRoom)
	for i := range rooms {
		deviceMap[rooms[i].Room] = rooms[i]
	}
	toReturn.roomCache = roomMap

	return toReturn
}
