package forwarding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
)

func Forward(e interface{}, url string) *nerr.E {
	start := time.Now()

	log.L.Debugf("Forwarding event %+v to %v", e, url)
	b, err := json.Marshal(e)
	if err != nil {
		return nerr.Translate(err).Addf("unable to forward event")
	}

	resp, err := http.Post(url, "appliciation/json", bytes.NewBuffer(b))
	if err != nil {
		return nerr.Translate(err).Addf("unable to forward event")
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nerr.Translate(err).Addf("failed to forward event. response status code: %v", resp.StatusCode)
		}
		return nerr.Create(fmt.Sprintf("failed to forward event. response status code: %v. response body: %s", resp.StatusCode, b), http.StatusText(resp.StatusCode))
	}

	log.L.Debugf("Successfully forwarded event. Took: %v", time.Since(start).Nanoseconds())
	return nil
}
