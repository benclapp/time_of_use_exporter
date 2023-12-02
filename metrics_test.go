package main

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestDescribeLocalizedTimezones(t *testing.T) {
	testCh := make(chan *prometheus.Desc)
	go describeLocalizedTimezones(testCh)

	observationCount := map[string]int{
		"minute":       0,
		"hour":         0,
		"day_of_week":  0,
		"day_of_month": 0,
		"month":        0,
	}

	for i := 0; i < len(observationCount); i++ {
		d := <-testCh
		_, after, ok := strings.Cut(d.String(), "tou_exporter_localized_")
		if !ok {
			t.Fatal(`Could not find "tou_exporter_localized_" in metric description`)
		}
		observationCount[strings.Split(after, `"`)[0]]++
	}

	select {
	case d := <-testCh:
		assert.Equal(t, nil, d, "Channel should be empty, more metric descriptors than expected")
	default:
	}

	for k, v := range observationCount {
		assert.Equal(t, 1, v, "observationCount for "+k+" should be exactly 1")
	}
}
