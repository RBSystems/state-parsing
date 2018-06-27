package jobs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/byuoitav/common/log"
)

type JobConfig struct {
	Name     string    `json:"name"`
	Triggers []Trigger `json:"triggers"`
	Enabled  bool      `json:"enabled"`
}

type Trigger struct {
	Kind  string      `json:"kind"`  // required for all
	At    string      `json:"at"`    // required for 'time'
	Every string      `json:"every"` // required for 'interval'
	Match MatchConfig `json:"match"` // required for 'event'
}

type runner struct {
	Job          Job
	Config       JobConfig
	Trigger      Trigger
	TriggerIndex int
}

var jobConfigs []JobConfig

func init() {
	// parse configuration
	path := os.Getenv("JOB_CONFIG_LOCATION")
	if len(path) < 1 {
		path = "./config.json"
	}
	log.L.Infof("Parsing job configuration from: %s", path)

	// get path for scripts
	scriptPath := os.Getenv("JOB_SCRIPTS_PATH")
	if len(scriptPath) < 1 {
		scriptPath = "./scripts/" // default script path
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.L.Fatalf("failed to read job configuration: %s", err)
	}

	err = json.Unmarshal(b, &jobConfigs)
	if err != nil {
		log.L.Fatalf("failed to parse job configuration: %s", err)
	}

	// validate all jobs exist, create the script jobs
	for _, job := range jobConfigs {
		if !job.Enabled {
			continue
		}

		// check if job exists
		isValid := false
		for name, _ := range Jobs {
			if strings.EqualFold(job.Name, name) {
				isValid = true
				break
			}
		}

		// if it isn't valid, then check if it's a valid script path
		if !isValid {
			if _, err := os.Stat(scriptPath + job.Name); err != nil {
				log.L.Fatalf("job '%s' doesn't exist, and doesn't have a script that matches its name.", job.Name)
			}

			// add the job for this script to the jobs map
			Jobs[job.Name] = &ScriptJob{Path: scriptPath + job.Name}
		}
	}

	// check number of active jobs
	activeJobs := 0
	for _, job := range jobConfigs {
		if !job.Enabled {
			continue
		}

		if len(job.Triggers) == 0 {
			log.L.Warnf("job %s has no triggers, so it won't be run.", job.Name)
			continue
		}
		activeJobs++
	}

	if activeJobs == 0 {
		log.L.Warnf("no active jobs. quitting scheduler, and just forwarding events.")
	} else {
		log.L.Infof("Scheduling %v jobs.", activeJobs)
	}
}

func StartJobScheduler() {
	go startEventMatcher()

	for _, job := range jobConfigs {
		for i, trigger := range job.Triggers {
			runnable := &runner{
				Job:          Jobs[job.Name],
				Config:       job,
				Trigger:      trigger,
				TriggerIndex: i,
			}

			if strings.EqualFold(trigger.Kind, "daily") {
				go runnable.runDaily()
			} else if strings.EqualFold(trigger.Kind, "interval") {
				go runnable.runInterval()
			} else if strings.EqualFold(trigger.Kind, "match") {
				runnable.addMatch()
			}
		}
	}
}

func (r *runner) runJob() {
	actions := r.Job.Run()
	// do something with the actions
	log.L.Debugf("actions: %v", actions)
}

func (r *runner) runDaily() {
	tmpDate := fmt.Sprintf("2006-01-02T%s", r.Trigger.At)
	runTime, err := time.Parse(time.RFC3339, tmpDate)
	runTime = runTime.UTC()
	if err != nil {
		log.L.Warnf("unable to parse time '%s' to execute job %s daily. error: %s", r.Trigger.At, r.Config.Name, err)
		return
	}

	log.L.Infof("[%s|%v] Running daily at %s", r.Config.Name, r.TriggerIndex, runTime.Format("15:04:05 MST"))

	// figure out how long until next run
	now := time.Now()
	until := time.Until(time.Date(now.Year(), now.Month(), now.Day(), runTime.Hour(), runTime.Minute(), runTime.Second(), 0, runTime.Location()))
	if until < 0 {
		until = 24*time.Hour + until
	}

	log.L.Debugf("[%s|%v] Time to next run: %v", r.Config.Name, r.TriggerIndex, until)
	timer := time.NewTimer(until)

	for {
		<-timer.C
		log.L.Infof("[%s|%v] Running job...", r.Config.Name, r.TriggerIndex)
		r.runJob()

		timer.Reset(24 * time.Hour)
	}
}

func (r *runner) runInterval() {
	interval, err := time.ParseDuration(r.Trigger.Every)
	if err != nil {
		log.L.Warnf("unable to parse duration '%s' to execute job %s on an interval. error: %s", r.Trigger.Every, r.Config.Name, err)
		return
	}

	log.L.Infof("[%s|%v] Running every %v", r.Config.Name, r.TriggerIndex, interval)

	ticker := time.NewTicker(interval)
	for range ticker.C {
		log.L.Infof("[%s|%v] Running job...", r.Config.Name, r.TriggerIndex)
		r.runJob()
	}
}

func (r *runner) addMatch() {
	// initialize the match
	// TODO validate at least one regex is created
	if len(r.Trigger.Match.Hostname) > 0 {
		r.Trigger.Match.Regex.Hostname = regexp.MustCompile(r.Trigger.Match.Hostname)
	}

	if len(r.Trigger.Match.Timestamp) > 0 {
		r.Trigger.Match.Regex.Timestamp = regexp.MustCompile(r.Trigger.Match.Timestamp)
	}

	if len(r.Trigger.Match.LocalEnvironment) > 0 {
		r.Trigger.Match.Regex.LocalEnvironment = regexp.MustCompile(r.Trigger.Match.LocalEnvironment)
	}

	if len(r.Trigger.Match.Building) > 0 {
		r.Trigger.Match.Regex.Building = regexp.MustCompile(r.Trigger.Match.Building)
	}

	if len(r.Trigger.Match.Room) > 0 {
		r.Trigger.Match.Regex.Room = regexp.MustCompile(r.Trigger.Match.Room)
	}

	if len(r.Trigger.Match.Event.Type) > 0 {
		r.Trigger.Match.Regex.Event.Type = regexp.MustCompile(r.Trigger.Match.Event.Type)
	}

	if len(r.Trigger.Match.Event.Requestor) > 0 {
		r.Trigger.Match.Regex.Event.Requestor = regexp.MustCompile(r.Trigger.Match.Event.Requestor)
	}

	if len(r.Trigger.Match.Event.EventCause) > 0 {
		r.Trigger.Match.Regex.Event.EventCause = regexp.MustCompile(r.Trigger.Match.Event.EventCause)
	}

	if len(r.Trigger.Match.Event.Device) > 0 {
		r.Trigger.Match.Regex.Event.Device = regexp.MustCompile(r.Trigger.Match.Event.Device)
	}

	if len(r.Trigger.Match.Event.EventInfoKey) > 0 {
		r.Trigger.Match.Regex.Event.EventInfoKey = regexp.MustCompile(r.Trigger.Match.Event.EventInfoKey)
	}

	if len(r.Trigger.Match.Event.EventInfoValue) > 0 {
		r.Trigger.Match.Regex.Event.EventInfoValue = regexp.MustCompile(r.Trigger.Match.Event.EventInfoValue)
	}

	addMatchJob(r)
}
