package mom

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/state-parsing/alerts/base"
)

var momAlertURL string

func init() {
	momAlertURL = os.Getenv("MOM_ALERT_URL")

	if len(momAlertURL) == 0 {
		log.L.Fatalf("MOM_ALERT_URL not set.")
	}
}

type MomNotificationEngine struct {
}

func (m *MomNotificationEngine) SendNotifications(alerts []base.Alert) ([]base.AlertReport, error) {
	if len(alerts) == 0 {
		return []base.AlertReport{}, nil
	}

	log.L.Infof("Sending mom alerts...")
	reportChan := make(chan base.AlertReport)

	// send all the alerts
	for _, alert := range alerts {
		go sendMomAlert(alert, reportChan)
	}

	// collect all the reports
	var reports []base.AlertReport
	for range alerts {
		reports = append(reports, <-reportChan)
	}

	return reports, nil
}

func sendMomAlert(alert base.Alert, reportChan chan base.AlertReport) {
	log.L.Debugf("Sending mom alert for %s", alert.Device)

	// init the report to send to the channel
	report := base.AlertReport{
		Alert:   alert,
		Success: false,
	}
	defer func() {
		reportChan <- report
	}()

	// build the request
	req, err := http.NewRequest(http.MethodPost, momAlertURL, bytes.NewReader(alert.Content))
	if err != nil {
		report.Message = fmt.Sprintf("failed to build request: %s", err)
		return
	}
	req.Header.Add("content-type", "application/json")

	// execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		report.Message = fmt.Sprintf("failed to send request: %s", err)
		return
	}

	// read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		report.Message = fmt.Sprintf("unable to read response body: %s", err)
		return
	}

	// check response status code
	if resp.StatusCode/100 != 2 {
		report.Message = fmt.Sprintf("non-200 response received: %v. body: %s", resp.StatusCode, body)
		return
	}

	// success
	report.Success = true
	report.Message = time.Now().Format(time.RFC3339)
}
