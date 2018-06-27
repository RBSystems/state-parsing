package forwarding

import (
	"log"

	"github.com/fatih/color"
)

func startAggregator(incomingChannel <-chan StateDistribution, tickerChan <-chan int, hostname string) {

	state := make(map[string]interface{})

	//start our timer
	for {
		select {
		case val, ok := <-incomingChannel:
			if !ok {
				//channel is closed
				color.Set(color.FgRed)
				log.Printf("[%s] channel closed, exiting", hostname)
				color.Unset()
				return
			}
			//update our map
			state[val.Key] = val.Value

		case _, ok := <-tickerChan:
			if !ok {
				color.Set(color.FgRed)
				log.Printf("[%s] channel closed, exiting", hostname)
				color.Unset()
				return
			}
			//we package up what we have and ship it downstreamS
			sendDownstream(state)

			//clear the map
			state = make(map[string]interface{})
		}
	}
}

func sendDownstream(state map[string]interface{}) {

}
