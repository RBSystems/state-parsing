package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/state-parsing/actions/action"
)

const slackurl = "https://hooks.slack.com/services/"

type SlackAction struct {
	ChannelIdentifier string
}

func (s *SlackAction) Execute(a action.Action) action.Result {
	log.L.Infof("Executing slack action for %v", a.Device)

	result := action.Result{
		Action: a,
	}

	var reqBody []byte
	var err error

	switch v := a.Content.(type) {
	case []byte:
		reqBody = v
	case Alert:
		// marshal the request
		reqBody, err = json.Marshal(v)
		if err != nil {
			result.Error = nerr.Translate(err).Addf("failed to unmarshal slack alert")
			return result
		}
	default:
		result.Error = nerr.Create("action content was not a slack alert.", reflect.TypeOf("").String())
		return result
	}

	// pretty simple, just a post, the only thing that could be an issue is the proxies
	proxyUrl, err := url.Parse(os.Getenv("PROXY_ADDR"))
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}

	req, err := http.NewRequest(http.MethodPost, slackurl+s.ChannelIdentifier, bytes.NewReader(reqBody))
	if err != nil {
		result.Error = nerr.Translate(err).Addf("failed to build slack alert request")
		return result
	}
	req.Header.Add("content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = nerr.Translate(err).Addf("failed to send slack alert")
		return result
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result.Error = nerr.Translate(err).Addf("could not read response body after sending slack alert: %s", err)
		return result
	}

	if resp.StatusCode/100 != 2 {
		result.Error = nerr.Create(fmt.Sprintf("non-200 response recieved (code: %v). body: %s", resp.StatusCode, b), reflect.TypeOf(resp).String())
		return result
	}

	//it worked
	log.L.Infof("Successfully sent slack alert for %s.", a.Device)
	return result
}
