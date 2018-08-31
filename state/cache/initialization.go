package cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/state-parser/elk"
	"github.com/byuoitav/state-parser/state/statedefinition"
)

const MAX_SIZE = 10000

func InitializeCaches() {
	Caches = make(map[string]*Cache)

	defRoomIndx, defDevIndx := getIndexesByType(DEFAULT)
	//dmpsRoomIndx, dmpsDevIndx := getIndexesByType(DMPS)

	//get DEFAULT devices
	defaultDevs, err := GetStaticDevices(defDevIndx)
	if err != nil {
		log.L.Errorf(err.Addf("Couldn't get information for default device cache").Error())
	}

	//get DEFAULT rooms
	defaultRooms, err := GetStaticRooms(defRoomIndx)
	if err != nil {
		log.L.Errorf(err.Addf("Couldn't get information for default room cache").Error())
	}

	cache := makeCache(defaultDevs, defaultRooms)
	Caches[DEFAULT] = &cache
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

func GetStaticDevices(index string) ([]statedefinition.StaticDevice, *nerr.E) {
	query := elk.GenericQuery{
		Size: MAX_SIZE,
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
		Size: MAX_SIZE,
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
	toReturn := memorycache{
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
		roomMap[rooms[i].Room] = rooms[i]
	}
	toReturn.roomCache = roomMap

	return &toReturn
}
