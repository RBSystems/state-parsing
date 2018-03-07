package jobs

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/byuoitav/state-parsing/alerts"
	"github.com/byuoitav/state-parsing/alerts/base"
	"github.com/byuoitav/state-parsing/alerts/device"
	"github.com/byuoitav/state-parsing/common"
	"github.com/fatih/color"
)

type Orchestrator struct {
	Config []common.Configuration
	Jobs   []Job
}

type Job struct {
	Config   common.Configuration
	StopChan chan string
}

func (j *Job) Run(c common.Configuration) {
	log.Printf("[%v] Setting configuration.", c.Name)
	j.Config = c
	if j.Config.WaitForComplete {
		log.Printf("[%v] Running on an interval.", c.Name)
		j.runIntervalTask()
	} else {
		log.Printf("[%v] Running on an schedule.", c.Name)
		j.runScheduledTask()
	}
}

func (j *Job) runIntervalTask() {
	//run forever until a stop message is receieved
	for true {
		timer := time.NewTimer(time.Second * time.Duration(j.Config.Interval))

		select {
		case <-timer.C:
			color.Set(color.FgGreen)
			log.Printf("[%v] Starting run...", j.Config.Name)
			color.Unset()
			startTime := time.Now()
			j.execute()
			elapsed := time.Since(startTime)

			color.Set(color.FgGreen)
			log.Printf("[%v] Done.", j.Config.Name)
			log.Printf("[%v] Execution took %s.", j.Config.Name, elapsed)
			color.Unset()

		case <-j.StopChan:
			color.Set(color.FgHiRed)
			log.Printf("[%v] Stop message received. Stopping.", j.Config.Name)
			color.Unset()
			return
		}

	}
}

func (j *Job) runScheduledTask() {
	//run forever until a stop message is receieved

	//start a ticker, as we're running at the same schedule
	ticker := time.NewTicker(time.Second * time.Duration(j.Config.Interval))
	for true {

		select {
		case <-ticker.C:
			color.Set(color.FgGreen)
			log.Printf("[%v] Starting run...", j.Config.Name)
			color.Unset()

			//if we need to run concurrently we can just execute this on a go routine
			startTime := time.Now()
			j.execute()
			elapsed := time.Since(startTime)

			color.Set(color.FgGreen)
			log.Printf("[%v] Done.", j.Config.Name)
			log.Printf("[%v] Execution took %s.", j.Config.Name, elapsed)
			color.Unset()
		case <-j.StopChan:
			color.Set(color.FgHiRed)
			log.Printf("[%v] Stop message received. Stopping.", j.Config.Name)
			color.Unset()
			return
		}

	}
}

func (j *Job) execute() {

	log.Printf(color.HiGreenString("[%v] Starting run.", j.Config.Name))
	startTime := time.Now()
	switch j.Config.Type {

	case "script":
		j.executeScript()
	case "alert-factory":
		j.executeAlertFactory()
	default:
		log.Printf(color.HiRedString("[%v] no type associated with: %v", j.Config.Name, j.Config.Type))
	}
	log.Printf(color.HiGreenString("'[%v] Time Elapsed: %v. ", j.Config.Name, time.Since(startTime)))
	log.Printf(color.HiGreenString("'[%v] Done. ", j.Config.Name))
}

func (j *Job) executeScript() {
	//find the script, and run it
	var command string
	if len(os.Getenv("EVENT_PARSING_SCRIPTS_PATH")) < 1 {
		command = fmt.Sprintf("./scripts/%v.py", j.Config.Name)
	} else {
		command = fmt.Sprintf("%s/%v.py", os.Getenv("EVENT_PARSING_SCRIPTS_PATH"), j.Config.Name)
	}

	cmd := exec.Command(command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	color.Set(color.FgHiGreen)
	log.Printf("[%v] Executing script", j.Config.Name)
	color.Unset()

	err := cmd.Run()
	if err != nil {

		color.Set(color.FgHiRed)
		log.Printf("[%v]Error: %v", j.Config.Name, err.Error())
		color.Unset()
	}
	return

}

func (j *Job) executeAlertFactory() {
	log.Printf(color.HiRedString("[%v] Starting Factory run...", j.Config.Name))
	factory, ok := alerts.GetAlertFactory(j.Config.Name)
	if !ok {
		log.Printf(color.HiRedString("[%v]Error: No alert factory found for %v", j.Config.Name, j.Config.Name))
		return
	}

	alertsToSend, err := factory.Run(1)
	if err != nil {
		log.Printf(color.HiRedString("[%v]error: %v", j.Config.Name, err.Error()))
	}

	reports := []base.AlertReport{}
	engines := alerts.GetNotificationEngines()
	log.Printf(color.HiGreenString("'[%v] Sending notifications...", j.Config.Name))

	for k, v := range alertsToSend {
		reps, err := engines[k].SendNotifications(v)
		if err != nil {
			log.Printf(color.HiRedString("Issue sending the %v notifications. Error: %v", k, err.Error()))
		}
		reports = append(reports, reps...)
	}

	log.Printf(color.HiGreenString("'[%v] Marking Alert as sent.", j.Config.Name))
	//now we mark the reports as sent
	device.MarkLastAlertSent(reports)
}

func (o *Orchestrator) Start() {
	log.Printf("Starting orchestrator")
	config, err := common.GetConfiguration()
	o.Config = config
	if err != nil {
		log.Printf("Could not get configuration: %v", err.Error())
		return
	}

	for _, c := range o.Config {
		if c.Enabled {
			log.Printf("Starting to job for %v", c.Name)
			stopChan := make(chan string, 1)
			j := Job{StopChan: stopChan}
			go j.Run(c)

			o.Jobs = append(o.Jobs, j)
		}
	}
}
