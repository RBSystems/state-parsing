package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/byuoitav/common/log"
)

var apiAddr, username, password string

func init() {
	apiAddr = os.Getenv("ELK_DIRECT_ADDRESS")
	username = os.Getenv("ELK_SA_USERNAME")
	password = os.Getenv("ELK_SA_PASSWORD")

	if len(apiAddr) == 0 || len(username) == 0 || len(password) == 0 {
		log.L.Fatalf("ELASTIC_API_EVENTS, ELK_SA_USERNAME, or ELK_SA_PASSWORD is not set.")
	}
}

func MakeELKRequest(method, endpoint string, query interface{}) ([]byte, error) {
	// format whole address
	addr := fmt.Sprintf("%s%s", apiAddr, endpoint)
	log.L.Infof("Making ELK request against: %s", addr)

	// marshal the request
	reqBody, err := json.Marshal(query)
	if err != nil {
		log.L.Warnf("failed to marshal query: %s", err)
	}

	// create the request
	req, err := http.NewRequest(method, addr, bytes.NewReader(reqBody))
	if err != nil {
		log.L.Warnf("there was a problem forming the request: %s", err)
		return []byte{}, err
	}

	// add auth
	req.SetBasicAuth(username, password)

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.L.Warnf("there was a problem making the request: %s", err)
		return []byte{}, err
	}

	// read the resp
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.L.Warnf("there was a problem making the request: %s", err)
		return []byte{}, err
	}

	// check resp code
	if resp.StatusCode/100 != 2 {
		log.L.Warnf("non 200 reponse code received. code: %v, body: %s", resp.StatusCode, respBody)
		return []byte{}, err // TODO should include the above message
	}

	return respBody, nil
}
