package statedefinition

import "time"

type Alert struct {
	AlertSent time.Time `json:"alert-sent,omitempty"`
	Alerting  bool      `json:"alerting,omitempty"`
	Message   string    `json:"message,omitempty"`
}

func compareAlerts(base, new map[string]Alert, changes bool) (diff map[string]Alert, merged map[string]Alert, changes bool) {
	for k, v := range new {
		basev, ok := base[k]
		if !ok {
			changes = true
			base[k] = v
			diff[k] = v
		}
		new, tempChanges := compareAlert(basev, v)
		if tempChanges {
			changes = true
			base[k] = new
			diff[k] = new
		}
	}
	return
}

func compareAlert(base, new Alert) (after alert, changes bool) {
	if base.Alerting != new.Alerting || base.AlertSent.Equal(new.AlertSent) || base.Message != new.Message {
		after = new
		changes = true
		return
	}
}
