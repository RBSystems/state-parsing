package common

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/byuoitav/state-parsing/logger"
	"github.com/fatih/color"
)

var apiAddr, username, password string

func init() {
	apiAddr = os.Getenv("ELASTIC_API_EVENTS")
	username = os.Getenv("ELK_SA_USERNAME")
	password = os.Getenv("ELK_SA_PASSWORD")

	if len(apiAddr) == 0 || len(username) == 0 || len(password) == 0 {
		log.Fatalf(color.HiRedString("ELASTIC_API_EVENTS, ELK_SA_USERNAME, or ELK_SA_PASSWORD is not set."))
	}
}

type ElkQuery struct {
	Method   string
	Endpoint string
	Query    string
}

func (q *ElkQuery) MakeELKRequest(logLevel int, caller string) (int, []byte, error) {
	l := logger.Logger{
		LogLevel: logLevel,
		Name:     caller,
	}

	// format whole address
	addr := fmt.Sprintf("%s%s", apiAddr, q.Endpoint)
	l.I("Making ELK request against %s", addr)

	// create the request
	req, err := http.NewRequest(q.Method, addr, bytes.NewReader([]byte(q.Query)))
	if err != nil {
		l.E("there was a problem forming the request: %s", err)
		return 0, []byte{}, err
	}

	// add auth
	req.SetBasicAuth(username, password)

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		l.E("there was a problem making the request: %s", err)
		return 0, []byte{}, err
	}

	// read the resp
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.E("there was a problem reading the response body: %s", err)
		return 0, []byte{}, err
	}

	// check resp code
	if resp.StatusCode/100 != 2 {
		l.E("non 200 response code received. code: %v, body: %s", resp.StatusCode, b)
	}

	return resp.StatusCode, b, nil
}
