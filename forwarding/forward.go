package forwarding

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/state-parser/elk"
)

var QUEUE_SIZE = 100000

//forwarding interval in milliseconds
var FORWARD_INTERVAL = 1000

type forwardData struct {
	Data   []byte
	Header elk.UpdateHeader
}

var forwardingqueue chan forwardData

func init() {
	queue := os.Getenv("QUEUE_SIZE")
	if len(queue) > 1 {
		val, err := strconv.Atoi(queue)
		if err != nil {
			log.L.Warnf("Invalid Queue size provided. Must be an integer")
		} else {
			QUEUE_SIZE = val
		}
	}

	ticker := time.NewTicker(time.Duration(FORWARD_INTERVAL) * time.Millisecond)
	forwardingqueue = make(chan forwardData, QUEUE_SIZE)

	//start our worker
	go forwardWorker(forwardingqueue, ticker)
}

// Forward makes a post request with <data> to <url>
// 2018-08-10 JB: We're running into an isssue where there aren't enough sockets available on the host for this to work, so we're moving to a forwarding queue.
func Forward(data interface{}, header elk.UpdateHeader) *nerr.E {
	if len(header.Index) == 0 {
		return nerr.Create("Index must not be null in update header", reflect.TypeOf("").String())
	}

	b, err := json.Marshal(data)
	if err != nil {
		return nerr.Translate(err).Addf("unable to forward event")
	}
	//we'll format it to be used in the bulk update endpoint, where URL is the index/type it needs to be.

	forwardingqueue <- forwardData{Data: b, Header: header}
	return nil
}

//this is the aggregator that will take in the two elements and add them to the payload that will eventually be pushed up to the bulk endpoint.
func forwardWorker(queue chan forwardData, dispatch *time.Ticker) {

	payload := []byte{}
	header := make(map[string]elk.UpdateHeader)

	for {
		select {
		case val := <-queue:

			//we do this to avoid allocating a new map everytime this gets called
			header["index"] = val.Header
			payload = append(payload, marshal(header, val.Data)...)
		case <-dispatch.C:
			//send it off and create a new payload so we don't have multiple handlles on the deal
			send(payload)
			payload = []byte{}
		}
	}
}

func marshal(header map[string]elk.UpdateHeader, body []byte) []byte {

	b, err := json.Marshal(header)
	if err != nil {
		log.L.Warnf("[dispatcher] there was a problem marshalling a line: %v", header)
		return []byte{}
	}

	b = append(b, '\n')
	b = append(b, body...)
	return append(b, '\n')
}

func send(payload []byte) {
	log.L.Infof("[forwarder]forwarding events.")

	//send the request
	req, err := http.NewRequest("POST", elk.APIAddr+"/_bulk", bytes.NewReader(payload))
	if err != nil {
		log.L.Warnf("[forwarder] there was a problem building the request: %v", err)
	}

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.L.Warnf("[forwarder] there was a problem sending the request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.L.Warnf("[forwarder] there was a non-200 respose: %v", resp.StatusCode)
		respBody, _ := ioutil.ReadAll(resp.Body)
		log.L.Warnf("[forwarder] error: %s", respBody)
		return
	}

	log.L.Infof("[forwarder] Done forwarding events.")
}

// 2018-08-13 JB: Leaving in here for legacy purposes, Delete if we're not using in a bit
/*
func forward(data []byte, url string) bool {
	start := time.Now()

	log.L.Debugf("Forwarding event %s to %v", data, url)

	resp, err := http.Post(url, "appliciation/json", bytes.NewBuffer(data))
	if err != nil {
		log.L.Warnf("Unable to forward event: %v", err.Error())
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.L.Warnf("Failed to forward event. response status code: %v. Error: %v", resp.StatusCode, err.Error())
			return false
		}
		log.L.Warnf("failed to forward event. response status code: %v. response body: %s", http.StatusText(resp.StatusCode), b)
		return false
	}
	log.L.Debugf("Successfully forwarded event. Took: %v", time.Since(start).Nanoseconds())
	return true
}
*/
