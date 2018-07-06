package elk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
)

const (
	DEVICE_INDEX = "oit-static-av-devices"
	ROOM_INDEX   = "oit-static-av-rooms"

	ALERTING_TRUE  = 1
	ALERTING_FALSE = 0
	POWER_STANDBY  = "standby"
	POWER_ON       = "on"
)

var apiAddr, username, password string

func init() {
	apiAddr = os.Getenv("ELK_DIRECT_ADDRESS") // or should this be ELK_ADDR?
	username = os.Getenv("ELK_SA_USERNAME")
	password = os.Getenv("ELK_SA_PASSWORD")

	if len(apiAddr) == 0 || len(username) == 0 || len(password) == 0 {
		log.L.Fatalf("ELASTIC_API_EVENTS, ELK_SA_USERNAME, or ELK_SA_PASSWORD is not set.")
	}
}

func MakeELKRequest(method, endpoint string, body interface{}) ([]byte, *nerr.E) {
	// format whole address
	addr := fmt.Sprintf("%s%s", apiAddr, endpoint)
	log.L.Debugf("Making ELK request against: %s", addr)

	var reqBody []byte
	var err error

	// marshal request if not already an array of bytes
	switch v := body.(type) {
	case []byte:
		reqBody = v
	default:
		// marshal the request
		reqBody, err = json.Marshal(v)
		if err != nil {
			return []byte{}, nerr.Translate(err)
		}
	}

	// create the request
	req, err := http.NewRequest(method, addr, bytes.NewReader(reqBody))
	if err != nil {
		return []byte{}, nerr.Translate(err)
	}

	// add auth
	req.SetBasicAuth(username, password)

	// add headers
	if method == http.MethodPost || method == http.MethodPut {
		req.Header.Add("content-type", "application/json")
	}

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, nerr.Translate(err)
	}
	defer resp.Body.Close()

	// read the resp
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, nerr.Translate(err)
	}

	// check resp code
	if resp.StatusCode/100 != 2 {
		msg := fmt.Sprintf("non 200 reponse code received. code: %v, body: %s", resp.StatusCode, respBody)
		return []byte{}, nerr.Create(msg, reflect.TypeOf(resp).String())
	}

	return respBody, nil
}