package forwarding

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/state-parsing/elk"
)

var dispatchChan chan string

//we use the sizeChannel to maintain the number of responses we expect for a given state update
var sizeChan chan int

var count int

func startDispatcher() {
	log.L.Infof("Starting dispatcher...")

	dispatchChan = make(chan string, 1000)
	sizeChan = make(chan int)

	expected := 0
	go func() {
		for {
			select {
			case _, _ = <-dispatchChan:
				//there needs to be some sort of "I don't have anything" marker - so we at least know to mark that routine as having sent something

			case newNum, ok := <-sizeChan:
				if !ok {
					log.L.Infof("Dispatcher number channel closed, exiting...")
				}
				expected = newNum
			}
		}
	}()
}

func dispatchLocalState(stateMap map[string]map[string]interface{}, mapType string) {
	if len(stateMap) < 1 {
		count++
		if count%10 == 0 {
			log.L.Infof("[dispatcher] No state to send.")
		}
		return
	}
	count = 0

	log.L.Infof("[dispatcher] Sending a state update...")

	//build our payload and send it off
	payload := []byte{}

	elkaddr := os.Getenv("ELK_DIRECT_ADDRESS")

	index := getIndexName(mapType)

	headerWrapper := make(map[string]elk.UpdateHeader)

	for k, v := range stateMap {
		recordType, err := getRecordType(k, mapType)
		if err != nil {
			//get our dev type split := strings.Split(k, "-") if len(split) < 3 {
			log.L.Warnf("[dispatcher] invalid hostname: %v", err)
			continue
		}

		//fill our meta data
		fillMeta(k, mapType, v)

		//build our first line
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
	req, err := http.NewRequest("POST", elkaddr+"/_bulk", bytes.NewReader(payload))
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

func fillDeviceMeta(name string, toFill map[string]interface{}) {
	split := strings.Split(name, "-")

	if len(split) != 3 {
		log.L.Warnf("[dispatcher] invalid hostname for device: %v", name)
		return
	}

	toFill["hostname"] = name
	toFill["room"] = split[0] + "-" + split[1]
	toFill["control"] = name
	toFill["view-dashboard"] = name
	toFill["suppress-notifications"] = name
	toFill["enable-notifications"] = name
	toFill["last-state-recieved"] = time.Now().Format(time.RFC3339)
}

func fillRoomMeta(name string, toFill map[string]interface{}) {
	split := strings.Split(name, "-")

	if len(split) != 2 {
		log.L.Warnf("[dispatcher] Invalid name for room: %v", name)
		return
	}

	toFill["enable-alerts"] = name
	toFill["suspend-alerts"] = name
	toFill["room"] = name
	toFill["view-alerts"] = name
	toFill["view-devices"] = name
	toFill["building"] = split[0]
	toFill["last-state-recieved"] = time.Now().Format(time.RFC3339)
}

// room record type is just 'room'
func getRoomRecordType(name string) (string, error) {
	split := strings.Split(name, "-")

	if len(split) != 2 {
		msg := fmt.Sprintf("[dispatcher] Invalid name for room: %v", name)
		log.L.Warn(msg)
		return "", errors.New(msg)
	}

	return "room", nil
}

var translationMap = map[string]string{
	"D":  "display",
	"CP": "control-processor",

	"DSP":   "digital-signal-processor",
	"PC":    "general-computer",
	"SW":    "video-switcher",
	"MICJK": "microphone-jack",
	"SP":    "scheduling-panel",
	"MIC":   "microphone",
	"DS":    "divider-sensor",
	"GW":    "gateway",
	"VIA":   "kramer-via",
	"HDMI":  "hdmi-input",
}

// device record type is determined usin the translation map
func getDeviceRecordType(name string) (string, error) {
	split := strings.Split(name, "-")
	if len(split) != 3 {
		msg := fmt.Sprintf("[dispatcher] Invalid hostname for device: %v", name)
		log.L.Warn(msg)
		return "", errors.New(msg)
	}

	for pos, char := range split[2] {
		if unicode.IsDigit(char) {
			val, ok := translationMap[split[2][:pos]]
			if !ok {
				msg := fmt.Sprintf("Invalid device type: %v", split[2][:pos])
				log.L.Warn(msg)
				return "", errors.New(msg)
			}
			return val, nil
		}
	}

	msg := fmt.Sprintf("no valid translation for :%v", split[2])
	log.L.Warn(msg)
	return "", errors.New(msg)
}

func getIndexName(mapType string) string {
	switch mapType {
	case "room":
		return os.Getenv("ELK_STATIC_ROOM_INDEX")
	case "device":
		return os.Getenv("ELK_STATIC_DEVICE_INDEX")
	}

	return ""
}
