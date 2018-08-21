package state

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/byuoitav/common/log"
)

const DEVICE = "device"
const ROOM = "room"

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

func GetIndexName(mapType string) string {
	switch mapType {
	case ROOM:
		return os.Getenv("ELK_STATIC_ROOM_INDEX")
	case DEVICE:
		return os.Getenv("ELK_STATIC_DEVICE_INDEX")
	}

	return ""
}
