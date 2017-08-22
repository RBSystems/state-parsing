package main

import (
	"sync"

	"github.com/byuoitav/state-parsing/jobs"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)

	o := jobs.Orchestrator{}
	o.Start()
	wg.Wait()
}
