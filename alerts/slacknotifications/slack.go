package slacknotifications

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/byuoitav/state-parsing/alerts/base"
	"github.com/fatih/color"
)

var url = "https://hooks.slack.com/services/"

//note don't forget to set the HTTP_PROXY or HTTPS_PROXY env variables if proxies are needed
type SlackNotificationEngine struct {
	ChannelIdentifier string
}

func (sn *SlackNotificationEngine) SendNotifications(alerts []base.Alert) ([]base.AlertReport, error) {
	log.Printf(color.HiGreenString("Sending slack notifications..."))

	//pretty simple, just a post, the only thing that could be an issue is the proxies
	report := []base.AlertReport{}

	for i := range alerts {
		log.Printf(color.HiGreenString("Sending for %v", alerts[i].Device))

		resp, err := http.Post(url+sn.ChannelIdentifier, "application/json", bytes.NewReader(alerts[i].Content))
		if err != nil {
			msg := fmt.Sprintf("Could not send request: %v", err.Error())
			log.Printf(color.HiRedString(msg))
			report = append(report, base.AlertReport{
				Alert:   alerts[i],
				Success: false,
				Message: msg,
			})
			continue
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			msg := fmt.Sprintf("Could not read response body: %v", err.Error())
			log.Printf(color.HiRedString(msg))
			report = append(report, base.AlertReport{
				Alert:   alerts[i],
				Success: false,
				Message: msg,
			})
			continue
		}

		if resp.StatusCode/100 != 2 {
			msg := fmt.Sprintf("Non-200 received: %v. Body: %s", resp.StatusCode, b)
			log.Printf(color.HiRedString(msg))
			report = append(report, base.AlertReport{
				Alert:   alerts[i],
				Success: false,
				Message: msg,
			})
			continue
		}

		//it worked
		log.Printf(color.HiGreenString("Success."))
		report = append(report, base.AlertReport{
			Alert:   alerts[i],
			Success: true,
			Message: time.Now().Format(time.RFC3339),
		})
	}

	return report, nil
}
