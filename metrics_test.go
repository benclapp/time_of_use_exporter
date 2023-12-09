package main

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

var observationCount = map[string]int{
	"minute":       0,
	"hour":         0,
	"day_of_week":  0,
	"day_of_month": 0,
	"month":        0,
}

func verifyMetricDescription(d *prometheus.Desc, t *testing.T) (bool, string) {
	_, after, ok := strings.Cut(d.String(), "tou_exporter_localized_")
	if !ok {
		slog.Error(`Could not find "tou_exporter_localized_" in metric description`, "desc", d)
		t.Fail()
		return false, ""
	}

	n := strings.Split(after, `"`)[0]
	if _, ok := observationCount[n]; ok {
		return true, n
	}
	return false, ""
}

func TestDescribeLocalizedTimezones(t *testing.T) {
	testCh := make(chan *prometheus.Desc)
	go describeLocalizedTimezones(testCh)

	var oc map[string]int = observationCount

	for i := 0; i < len(oc); i++ {
		ok, n := verifyMetricDescription(<-testCh, t)
		if ok {
			oc[n]++
		}
	}

	select {
	case d := <-testCh:
		assert.Equal(t, nil, d, "Channel should be empty, more metric descriptors than expected")
	default:
	}

	for k, v := range oc {
		assert.Equal(t, 1, v, k+" should have been observed exactly once")
	}
}

func TestCollectLocalizedTimezones(t *testing.T) {
	testCollectCh := make(chan prometheus.Metric)
	liveConfig = config{LocalizedTimezones: []string{
		"Pacific/Chatham", // UTC+13:45 - tests minute offsets too
	}}
	tTime := time.Date(2023, 1, 31, 20, 3, 4, 0, time.UTC)
	go collectLocalizedTimezones(testCollectCh, tTime)

	// var oc map[string]int = observationCount
	var oc = map[string]int{}
	for k, _ := range observationCount {
		oc[k] = 0
	}

	for i := 0; i < len(oc); i++ {
		m := <-testCollectCh
		ok, n := verifyMetricDescription(m.Desc(), t)
		if ok {
			oc[n]++

			var actual = &dto.Metric{}
			if err := m.Write(actual); err != nil {
				slog.Error("error writing metric", "err", err, "metric", m, "metricDesc", m.Desc())
				t.Fail()
			}
			actualValue := actual.GetGauge().GetValue()

			var labelMap = map[string]string{}
			for _, l := range actual.GetLabel() {
				labelMap[l.GetName()] = l.GetValue()
			}

			assert.Equal(t, "Pacific/Chatham", labelMap["tz"], "minute label should be Pacific/Chatham")
			switch n {
			case "minute":
				assert.Equal(t, float64(48), actualValue, "minute should be 48")
			case "hour":
				assert.Equal(t, float64(9), actualValue, "hour should be 9")
			case "day_of_week":
				assert.Equal(t, float64(3), actualValue, "day_of_week should be 3 (wednesday)")
				assert.Equal(t, "Wednesday", labelMap["day"], "day_of_week label should be Wednesday")
			case "day_of_month":
				assert.Equal(t, float64(1), actualValue, "day_of_month should be 1st")
			case "month":
				assert.Equal(t, float64(2), actualValue, "month should be 2 (feb)")
				assert.Equal(t, "February", labelMap["month"], "day_of_month label should be February")
			}
		}
	}

	select {
	case m := <-testCollectCh:
		assert.Equal(t, nil, m, "Channel should be empty, more metrics than expected")
	default:
	}

	for k, v := range oc {
		assert.Equal(t, 1, v, k+" should have been observed exactly once")
	}
}
