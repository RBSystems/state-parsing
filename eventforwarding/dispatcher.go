package eventforwarding

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/byuoitav/state-parsing/common"
	"github.com/fatih/color"
)

var dispatchChan chan string

//we use the sizeChannel to maintain the number of responses we expect for a given state update
var sizeChan chan int

var count int

func startDispatcher() {
	log.Printf("[Dispatcher] Starting dispatcher")

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
					color.Set(color.FgRed)
					log.Printf("[dispatch] number channel closed, exiting")
					color.Unset()
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
			color.Set(color.FgYellow)
			log.Printf("[Dispatcher] no state to send.")
			color.Unset()
		}
		return
	}
	count = 0
	color.Set(color.FgGreen)
	log.Printf("[Dispatcher] Sending a state update...")
	color.Unset()

	//build our payload and send it off
	payload := []byte{}

	elkaddr := os.Getenv("ELK_DIRECT_ADDRESS")

	index := getIndexName(mapType)

	headerWrapper := make(map[string]common.UpdateHeader)

	for k, v := range stateMap {

		recordType, err := getRecordType(k, mapType)
		if err != nil {
			//get our dev type split := strings.Split(k, "-") if len(split) < 3 {
			log.Printf("[dispatcher] Invalid hostname: %v", err.Error())
			continue
		}

		//fill our meta data
		fillMeta(k, mapType, v)

		//build our first line
		headerWrapper["update"] = common.UpdateHeader{ID: k, Type: recordType, Index: index}
		ub := common.UpdateBody{Doc: v, Upsert: true}

		b, err := json.Marshal(headerWrapper)
		if err != nil {
			color.Set(color.FgRed)
			log.Printf("[Dispatcher] There was a problem marshalling a line: %v", headerWrapper)
			color.Unset()
			continue
		}
		bb, err := json.Marshal(ub)
		if err != nil {
			color.Set(color.FgRed)
			log.Printf("[Dispatcher] There was a problem marshalling a line: %v", ub)
			color.Unset()
			continue
		}

		//add to our payload
		payload = append(payload, b...)
		payload = append(payload, '\n')
		payload = append(payload, bb...)
		payload = append(payload, '\n')

		color.Set(color.FgYellow)
		//	log.Printf("[Dispatcher] Added line for device %v.", k)
		color.Unset()
	}

	color.Set(color.FgGreen)
	log.Printf("[Dispatcher] Done adding lines.")
	log.Printf("[Dispatcher] %v devices getting updates....", len(stateMap))
	color.Unset()

	//log.Printf("\n%s", payload)

	//send the request
	req, err := http.NewRequest("POST", elkaddr+"/_bulk", bytes.NewReader(payload))
	if err != nil {
		color.Set(color.FgRed)
		log.Printf("[Dispatcher] There was a problem building the request: %v", err.Error())
		color.Unset()
	}

	req.SetBasicAuth(os.Getenv("ELK_SA_USERNAME"), os.Getenv("ELK_SA_PASSWORD"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		color.Set(color.FgRed)
		log.Printf("[Dispatcher] There was a problem sending the request: %v", err.Error())
		color.Unset()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		color.Set(color.FgRed)
		log.Printf("[Dispatcher] There was a non-200 respose: %v", resp.StatusCode)
		respBody, _ := ioutil.ReadAll(resp.Body)
		log.Printf("[Dispatcher] Error: %s", respBody)
		color.Unset()

		resp.Body.Close()
		return
	}
	color.Set(color.FgGreen)
	log.Printf("[Dispatcher] Done dispatching state.")
	color.Unset()
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
		log.Printf(color.HiRedString("[dispatcher] Invalid hostname for device: %v", name))
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
		log.Printf(color.HiRedString("[dispatcher] Invalid name for room: %v", name))
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

//room record type is just 'room'
func getRoomRecordType(name string) (string, error) {
	split := strings.Split(name, "-")

	if len(split) != 2 {
		msg := fmt.Sprintf("[dispatcher] Invalid name for room: %v", name)
		log.Printf(color.HiRedString(msg))
		return "", errors.New(msg)
	}
	return "room", nil
}

var translationMap = map[string]string{
	"D":  "display",
	"CP": "control-processor",

	"DSP": "digital-signal-processor",
	"PC":  "general-computer",
	"SW":  "video-switcher",
}

//device record type is determined usin the translation map
func getDeviceRecordType(name string) (string, error) {
	split := strings.Split(name, "-")
	if len(split) != 3 {
		msg := fmt.Sprintf("[dispatcher] Invalid hostname for device: %v", name)
		log.Printf(color.HiRedString(msg))
		return "", errors.New(msg)
	}
	for pos, char := range split[2] {
		if unicode.IsDigit(char) {
			val, ok := translationMap[split[2][:pos]]
			if !ok {
				msg := fmt.Sprintf("Invalid device type: %v", split[2][:pos])
				log.Printf(color.HiRedString(msg))
				return "", errors.New(msg)
			}
			return val, nil
		}
	}
	msg := fmt.Sprintf("no valid translation for :%v", split[2])
	log.Printf(color.HiRedString(msg))
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
