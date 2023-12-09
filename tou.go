package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func describeTOUMetric(tou timeOfUse) *prometheus.Desc {
	labels := map[string]string{"tz": "UTC"}
	if tou.Timezone != "" {
		labels["tz"] = tou.Timezone
	}
	for k, v := range tou.Labels {
		labels[k] = v
	}

	return prometheus.NewDesc(
		tou.Name,
		tou.Description,
		nil,
		labels,
	)
}

func calculateTOUValue(tou timeOfUse, now time.Time) float64 {
	for _, tw := range tou.TimeWindows {
		if isWithinTimeWindow(tw, now) {
			return tw.Value
		}
	}
	return tou.DefaultValue
}

func isWithinTimeWindow(tw timeWindow, now time.Time) bool {
	basedDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if now.Equal(basedDay.Add(tw.startDuration)) ||
		now.After(basedDay.Add(tw.startDuration)) &&
			now.Before(basedDay.Add(tw.endDuration)) {
		return true
	}
	return false
}
