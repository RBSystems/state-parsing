package forwarding

import "time"

var localTickerChan chan bool

//interval in milliseconds
func StartTicker(interval int) {

	localTickerChan = make(chan bool, 1)

	//run it local
	if runLocal == true {

		ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)

		for _ = range ticker.C {
			localTickerChan <- true
		}

	}
	//other wise we need to let people register to get ticks, and then we send them out
}
