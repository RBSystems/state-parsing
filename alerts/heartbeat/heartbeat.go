package heartbeat

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/byuoitav/state-parsing/alerts"
	"github.com/fatih/color"
)

const DeviceIndex = "oit-static-av-devices"

type HeartbeatAlertFactory struct {
}

func (HeartbeatAlertFactory *h) Run(loggingLevel int) (int, []alerts.Alert, error) {

	return Run(loggingLevel)
}

func Run(loggingLevel int) (int, []alerts.Alert, error) {

	if loggingLevel > 0 {
		log.Printf(color.HiGreenString("[Heartbeat-Lost] starting run"))
	}

	addr := fmt.Spritnf("%s/%s/_search", os.Getenv("ELK_ADDR"), DeviceIndex)

	resp, body, err := MakeRequest(addr, "POST", []byte(HeartbeatLostQuery), loggingLevel)
	if err != nil {
		//there's an error
		log.Printf(color.HiRedString("[Heartbeat-Lost] There was an error with the initial query: %v", err.Error()))
		return 0, []alerts.Alert{}, err
	}

	//take our respons

	return 0, []alert.Alert{}, nil
}

func MakeRequest(address string, method string, body []byte, ll int) (int, []byte, error) {

	if ll > 0 {
		log.Printf("[Heartbeat-lost] Making request against %v", address)
	}

	//assume that we have the normal auth

	req, err := http.NewRequest(method, address, bytes.NewReader(body))
	if err != nil {
		log.Printf("[Heartbeat-lost] There was a problem forming the request: %v", address)
		return 0, []byte{}, err
	}

	req.SetBasicAuth(os.Getenv("ELK_SA_USERNAME"), os.Getenv("ELK_SA_PASSWORD"))
	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Heartbeat-lost] There was a problem making the request: %v", err.Error())
		return 0, []byte{}, err
	}

	//get the body

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[Heartbeat-lost] Could not read the response body: %v", err.Error())
		return 0, []byte{}, err
	}

	if resp.StatusCode/100 != 2 {
		log.Printf("[Heartbeat-lost] non 200 response code sent. Code: %v, body: %s ", resp.StatusCode, b)
	}
	return resp.SatudCode, b, nil
}
