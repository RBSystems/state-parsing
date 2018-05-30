package common

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/byuoitav/state-parsing/logger"
	"github.com/fatih/color"
)

var apiAddr, username, password string

type ElkQuery struct {
	Method   string
	Endpoint string
	Query    string
}

func init() {
	apiAddr = os.Getenv("ELK_DIRECT_ADDRESS")
	username = os.Getenv("ELK_SA_USERNAME")
	password = os.Getenv("ELK_SA_PASSWORD")

	if len(apiAddr) == 0 || len(username) == 0 || len(password) == 0 {
		log.Fatalf(color.HiRedString("ELASTIC_API_EVENTS, ELK_SA_USERNAME, or ELK_SA_PASSWORD is not set."))
	}
}

func (q *ElkQuery) MakeELKRequest(logLevel int, caller string) ([]byte, error) {
	l := logger.New(caller, logLevel)

	// format whole address
	addr := fmt.Sprintf("%s%s", apiAddr, q.Endpoint)
	l.Info("Making ELK request against %s", addr)

	// create the request
	req, err := http.NewRequest(q.Method, addr, bytes.NewReader([]byte(q.Query)))
	if err != nil {
		l.Error("there was a problem forming the request: %s", err)
		return []byte{}, err
	}

	// add auth
	req.SetBasicAuth(username, password)

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		l.Error("there was a problem making the request: %s", err)
		return []byte{}, err
	}

	// read the resp
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Error("there was a problem reading the response body: %s", err)
		return []byte{}, err
	}

	// check resp code
	if resp.StatusCode/100 != 2 {
		msg := fmt.Sprintf("non 200 reponse code received. code: %v, body: %s", resp.StatusCode, b)
		l.Error(msg)

		return []byte{}, errors.New(msg)
	}

	return b, nil
}
