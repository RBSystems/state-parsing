package jobs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/common/events"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/byuoitav/state-parser/actions"
	"github.com/byuoitav/state-parser/elk"
	"github.com/byuoitav/state-parser/state"
)

// var (

// 	MaxWorkers = os.Getenv("MAX_WORKERS")

// 	MaxQueue = os.Getenv("MAX_QUEUE")

// 	runners       []*runner
// 	eventChan     chan elkreporting.ElkEvent
// 	heartbeatChan chan events.Event
// )

//JobPackage - moving these variables to a struct so we can have multiples running at the same time
type JobPackage struct {
	MaxWorkers    int
	MaxQueue      int
	Runners       []*runner
	EventChan     chan elkreporting.ElkEvent
	HeartbeatChan chan events.Event
}

type runner struct {
	Job          Job
	Config       JobConfig
	Trigger      Trigger
	TriggerIndex int
}

//SetUpJobPackage - set up a job package to run
func SetUpJobPackage(MaxWorkers string, MaxQueue string, JobConfigLocation string, JobScriptsPath string) JobPackage {
	//create the return objects
	var thisPackage JobPackage
	var err error

	// set defaults for max workers/queue
	if len(MaxWorkers) == 0 {
		MaxWorkers = "10"
	}
	if len(MaxQueue) == 0 {
		MaxQueue = "1000"
	}

	// validate max workers/queue are valid numbers
	thisPackage.MaxWorkers, err = strconv.Atoi(MaxWorkers)
	if err != nil {
		log.L.Fatalf("$MAX_WORKERS must be a number")
	}
	thisPackage.MaxQueue, err = strconv.Atoi(MaxQueue)
	if err != nil {
		log.L.Fatalf("$MAX_QUEUE must be a number")
	}

	// parse configuration
	path := JobConfigLocation //os.Getenv("JOB_CONFIG_LOCATION")
	if len(path) < 1 {
		path = "./config.json"
	}
	log.L.Infof("Parsing job configuration from: %s", path)

	// get path for scripts
	scriptPath := JobScriptsPath //os.Getenv("JOB_SCRIPTS_PATH")
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
		for name := range Jobs {
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
			thisPackage.Runners = append(thisPackage.Runners, runner)
		}
	}

	thisPackage.EventChan = make(chan elkreporting.ElkEvent, thisPackage.MaxQueue)
	thisPackage.HeartbeatChan = make(chan events.Event, thisPackage.MaxQueue)

	return thisPackage
}

// StartJobScheduler starts workers to run jobs, defined in the config.json file.
func (thisPackage *JobPackage) StartJobScheduler() {
	log.L.Infof("Starting job scheduler. Running %v jobs with %v workers with a max of %v events queued at once.",
		len(thisPackage.Runners), thisPackage.MaxWorkers, thisPackage.MaxQueue)

	wg := sync.WaitGroup{}

	// start action managers
	go actions.StartActionManagers()

	// start runners
	var matchRunners []*runner
	for _, runner := range thisPackage.Runners {
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
	for i := 0; i < thisPackage.MaxWorkers; i++ {
		log.L.Debugf("Starting event worker %v", i)
		wg.Add(1)

		go func() {
			for {
				select {
				case event := <-thisPackage.EventChan:
					// see if we need to execute any jobs from this event
					for i := range matchRunners {
						if matchRunners[i].doesEventMatch(&event) {
							go matchRunners[i].run(&event)
						}
					}

				case heartbeat := <-thisPackage.HeartbeatChan:
					// forward heartbeat
					go state.Forward(heartbeat, elk.UpdateHeader{
						Index: elk.GenerateIndexName(elk.OIT_AV_HEARTBEAT),
						Type:  "heartbeat",
					})
					go state.DistributeHeartbeat(heartbeat)
				}
			}
		}()
	}

	wg.Wait()
}

func (r *runner) run(context interface{}) {
	log.L.Debugf("[%s|%v] Running job...", r.Config.Name, r.TriggerIndex)
	actions.Execute(r.Job.Run(context))
	log.L.Debugf("[%s|%v] Finished.", r.Config.Name, r.TriggerIndex)
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
