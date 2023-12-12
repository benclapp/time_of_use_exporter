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
	start := time.Date(now.Year(), now.Month(), now.Day(), tw.startHour, tw.startMinute, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), tw.endHour, tw.endMinute, 0, 0, now.Location())
	if now.Equal(start) || now.After(start) && now.Before(end) {
		return true
	}
	return false
}
