package common

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/fatih/color"
)

var username, password string

func init() {
	username = os.Getenv("ELK_SA_USERNAME")
	password = os.Getenv("ELK_SA_PASSWORD")

	if len(username) == 0 || len(password) == 0 {
		log.Fatalf(color.HiRedString("ELK_SA_USERNAME or ELK_SA_PASSWORD is not set."))
	}
}

func MakeELKRequest(address, method string, body []byte, ll int, caller string) (int, []byte, error) {
	if ll > 0 {
		log.Printf("[%s] Making request against %v", caller, address)
	}

	// create the request
	req, err := http.NewRequest(method, address, bytes.NewReader(body))
	if err != nil {
		log.Printf("[%s] There was a problem forming the request: %v", caller, address)
		return 0, []byte{}, err
	}

	// set the auth; assume using basic auth
	req.SetBasicAuth(username, password)

	// create the client and make the request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[%s] There was a problem making the request: %v", caller, err)
		return 0, []byte{}, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[%s] Could not read the response body: %v", caller, err)
		return 0, []byte{}, err
	}

	if resp.StatusCode/100 != 2 {
		log.Printf("[%s] non 200 response code sent. Code: %v, body: %s ", caller, resp.StatusCode, b)
	}

	if ll > 2 {
		log.Printf(color.HiGreenString("[%s] Elk Request successful with statuscode %v", caller, resp.StatusCode))
	}

	return resp.StatusCode, b, nil
}
