package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/state-parser/elk"
)

var count int

func dispatchState(stateMap map[string]map[string]interface{}, mapType string) {
	if len(stateMap) < 1 {
		count++
		if count%10 == 0 {
			log.L.Infof("[dispatcher] No state to send.")
		}
		return
	}
	count = 0

	log.L.Infof("[dispatcher] Sending a state update...")

	// build our payload and send it off
	payload := []byte{}

	index := GetIndexName(mapType)

	headerWrapper := make(map[string]elk.UpdateHeader)

	for k, v := range stateMap {
		recordType, err := getRecordType(k, mapType)
		if err != nil {
			//get our dev type split := strings.Split(k, "-") if len(split) < 3 {
			log.L.Warnf("[dispatcher] invalid hostname: %v", err)
			continue
		}

		// fill our meta data
		fillMeta(k, mapType, v)

		// build our first line
		headerWrapper["update"] = elk.UpdateHeader{ID: k, Type: recordType, Index: index}
		ub := elk.UpdateBody{Doc: v, Upsert: true}

		b, err := json.Marshal(headerWrapper)
		if err != nil {
			log.L.Warnf("[dispatcher] there was a problem marshalling a line: %v", headerWrapper)
			continue
		}
		bb, err := json.Marshal(ub)
		if err != nil {
			log.L.Warnf("[dispatcher] there was a problem marshalling a line: %v", ub)
			continue
		}

		//add to our payload
		payload = append(payload, b...)
		payload = append(payload, '\n')
		payload = append(payload, bb...)
		payload = append(payload, '\n')
	}

	log.L.Debugf("[dispatcher] Done adding lines.")
	log.L.Debugf("[dispatcher] %v devices getting updates....", len(stateMap))
	log.L.Debugf("%s", payload)

	//send the request
	req, err := http.NewRequest("POST", elk.APIAddr+"/_bulk", bytes.NewReader(payload))
	if err != nil {
		log.L.Warnf("[dispatcher] there was a problem building the request: %v", err)
	}

	req.SetBasicAuth(os.Getenv("ELK_SA_USERNAME"), os.Getenv("ELK_SA_PASSWORD"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.L.Warnf("[dispatcher] there was a problem sending the request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.L.Warnf("[dispatcher] there was a non-200 respose: %v", resp.StatusCode)
		respBody, _ := ioutil.ReadAll(resp.Body)
		log.L.Warnf("[dispatcher] error: %s", respBody)

		resp.Body.Close()
		return
	}

	log.L.Infof("[dispatcher] Done dispatching state.")
}

func getRecordType(hostname, mapType string) (string, error) {
	switch mapType {
	case "room":
		return getRoomRecordType(hostname)
	case "device":
		return getDeviceRecordType(hostname)
	}

	return "", errors.New("Invalid mapType")
}

func fillMeta(name, mapType string, toFill map[string]interface{}) {
	switch mapType {
	case "room":
		fillRoomMeta(name, toFill)
		return
	case "device":
		fillDeviceMeta(name, toFill)
		return
	}
}
