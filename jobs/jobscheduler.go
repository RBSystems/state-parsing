package jobs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/byuoitav/salt-translator-service/elk"
	"github.com/byuoitav/state-parsing/actions"
	"github.com/byuoitav/state-parsing/forwarding"
)

var (
	// buffered channel to send events through
	EventChan     chan elkreporting.ElkEvent
	HeartbeatChan chan elk.Event

	// maximum number of workers
	MAX_WORKERS = os.Getenv("MAX_WORKERS")

	// maximum size to queue events, before making the request hang
	MAX_QUEUE = os.Getenv("MAX_QUEUE")

	// forwarding urls
	API_FORWARD       = os.Getenv("ELASTIC_API_EVENTS")
	HEARTBEAT_FORWARD = os.Getenv("ELASTIC_HEARTBEAT_EVENTS")

	// private stuff
	runners []*runner
)

type runner struct {
	Job          Job
	Config       JobConfig
	Trigger      Trigger
	TriggerIndex int
}

func init() {
	// make sure max queue and max workers size is set
	if len(MAX_WORKERS) == 0 || len(MAX_QUEUE) == 0 {
		log.L.Fatalf("must set $MAX_WORKERS and $MAX_QUEUE before running.")
	}

	// validate max workers/queue are valid numbers
	_, err := strconv.Atoi(MAX_WORKERS)
	if err != nil {
		log.L.Fatalf("$MAX_WORKERS must be a number")
	}
	_, err = strconv.Atoi(MAX_QUEUE)
	if err != nil {
		log.L.Fatalf("$MAX_Queue must be a number")
	}

	// validate forwarding urls exist
	if len(API_FORWARD) == 0 || len(HEARTBEAT_FORWARD) == 0 {
		log.L.Fatalf("$ELASTIC_API_EVENTS and $ELASTIC_HEARTBEAT_EVENTS must be set.")
	}
	log.L.Infof("\n\nForwarding URLs:\n\tAPI_FORWARD:\t\t%v\n\tHEARTBEAT_FORWARD:\t%v\n", API_FORWARD, HEARTBEAT_FORWARD)

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

	// read job configuration
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.L.Fatalf("failed to read job configuration: %s", err)
	}

	// unmarshal job config
	var jobConfigs []JobConfig
	err = json.Unmarshal(b, &jobConfigs)
	if err != nil {
		log.L.Fatalf("failed to parse job configuration: %s", err)
	}

	// validate all jobs exist, create the script jobs
	for _, config := range jobConfigs {
		if !config.Enabled {
			continue
		}

		// check if job exists
		isValid := false
		for name, _ := range Jobs {
			if strings.EqualFold(config.Name, name) {
				isValid = true
				break
			}
		}

		// if it isn't valid, then check if it's a valid script
		if !isValid {
			if _, err := os.Stat(scriptPath + config.Name); err != nil {
				log.L.Fatalf("job '%s' doesn't exist, and doesn't have a script that matches its name.", config.Name)
			}

			// TODO check if the job already exists, and just reuse that script job

			// add the job for this script to the jobs map
			Jobs[config.Name] = &ScriptJob{Path: scriptPath + config.Name}
		}

		// build a runner for each trigger
		for i, trigger := range config.Triggers {
			runner := &runner{
				Job:          Jobs[config.Name],
				Config:       config,
				Trigger:      trigger,
				TriggerIndex: i,
			}

			// build the regex if it's a match type
			if strings.EqualFold(runner.Trigger.Type, "match") {
				runner.buildMatchRegex()
			}

			log.L.Infof("Adding runner for job '%v', trigger #%v. Execution type: %v", runner.Config.Name, runner.TriggerIndex, runner.Trigger.Type)
			runners = append(runners, runner)
		}
	}
}

func StartJobScheduler() {
	maxWorkers, _ := strconv.Atoi(MAX_WORKERS)
	maxQueue, _ := strconv.Atoi(MAX_QUEUE)

	log.L.Infof("Starting job scheduler. Running %v jobs with %v workers with a max of %v events queued at once.", len(runners), maxWorkers, maxQueue)

	EventChan = make(chan elkreporting.ElkEvent, maxQueue)
	HeartbeatChan = make(chan elk.Event, maxQueue)

	// start runners
	var matchRunners []*runner
	for _, runner := range runners {
		switch runner.Trigger.Type {
		case "daily":
			go runner.runDaily()
		case "interval":
			go runner.runInterval()
		case "match":
			matchRunners = append(matchRunners, runner)
		default:
			log.L.Warnf("unknown trigger type '%v' for job %v|%v", runner.Trigger.Type, runner.Config.Name, runner.TriggerIndex)
		}
	}

	// start event workers
	for i := 0; i < maxWorkers; i++ {
		log.L.Infof("Starting event worker %v", i)

		go func() {
			for {
				select {
				case event := <-EventChan:
					log.L.Debugf("Received event: %+v", event)

					// forward to elk
					go forwarding.Forward(&event, API_FORWARD)
					go forwarding.DistributeEvent(&event)

					// see if we need to execute any jobs from this event
					for _, runner := range matchRunners {
						if runner.doesEventMatch(&event) {
							go runner.run(&event)
						}
					}

				case heartbeat := <-HeartbeatChan:
					log.L.Debugf("Received heartbeat: %+v", heartbeat)

					// forward to elk
					go forwarding.Forward(&heartbeat, HEARTBEAT_FORWARD)
					go forwarding.DistributeHeartbeat(&heartbeat)
				}
			}
		}()
	}
}

func (r *runner) run(context interface{}) {
	log.L.Infof("[%s|%v] Running job...", r.Config.Name, r.TriggerIndex)
	actions.Execute(r.Job.Run(context))
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
		r.run(nil)

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
		r.run(nil)
	}
}
