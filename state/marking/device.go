package marking

import (
	"fmt"
	"time"

	"github.com/byuoitav/common/log"

	"github.com/byuoitav/state-parser/state/cache"
	sd "github.com/byuoitav/state-parser/state/statedefinition"
)

var False = false

func ClearHeartbeatAlerts(deviceIDs []string) {

	device := sd.StaticDevice{
		UpdateTimes: make(map[string]time.Time),
	}
	device.Alerts["lost-heartbeat"] = sd.Alert{
		Alerting: false,
		Message:  fmt.Sprintf("Alert cleared at %s", time.Now().Format(time.RFC3339)),
	}

	device.UpdateTimes["alerts.lost-heartbeat"] = time.Now()

	//for now, this will have to change once we add more alerts
	device.Alerting = &False

	for _, id := range deviceIDs {
		device.DeviceID = id
		_, _, err := cache.GetCache(cache.DEFAULT).CheckAndStoreDevice(device)
		if err != nil {
			log.L.Errorf("Couldn't clear hearteat alerts for %v: %v", id, err.Error())
		}
	}
}
