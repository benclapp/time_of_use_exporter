package main

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

type describeTOUTestCase struct {
	inputTou     timeOfUse
	expectedDesc *prometheus.Desc
}

func TestDescribeTOUMetric(t *testing.T) {
	tcs := map[string]describeTOUTestCase{
		"basic": {
			inputTou: timeOfUse{
				Name:        "basic",
				Description: "basic description",
			},
			expectedDesc: prometheus.NewDesc(
				"basic", "basic description", nil,
				map[string]string{"tz": "UTC"},
			),
		},
		"custom timezone": {
			inputTou: timeOfUse{
				Name:        "custom_timezone",
				Description: "custom timezone description",
				Timezone:    "Pacific/Auckland",
			},
			expectedDesc: prometheus.NewDesc(
				"custom_timezone", "custom timezone description",
				nil, map[string]string{"tz": "Pacific/Auckland"},
			),
		},
		"custom labels": {
			inputTou: timeOfUse{
				Name:        "custom_labels",
				Description: "custom labels description",
				Labels: map[string]string{
					"foo": "bar",
				},
			},
			expectedDesc: prometheus.NewDesc(
				"custom_labels", "custom labels description",
				nil, map[string]string{"tz": "UTC", "foo": "bar"},
			),
		},
	}

	for name, tc := range tcs {
		v := describeTOUMetric(tc.inputTou)
		assert.Equal(t, tc.expectedDesc, v, "test case %s failed", name)
	}
}

func TestCalculateTOUValue(t *testing.T) {
	assert.Equal(t, float64(123), calculateTOUValue(
		timeOfUse{
			Name:         "test",
			Description:  "test description",
			DefaultValue: 123,
			TimeWindows: []timeWindow{{
				Value:         456,
				Start:         "11:59",
				End:           "12:01",
				startDuration: time.Duration(11*time.Hour + 59*time.Minute),
				endDuration:   time.Duration(12*time.Hour + 1*time.Minute),
			}},
		},
		time.Date(2023, 12, 1, 12, 1, 0, 0, time.UTC),
	), "should return default value")
	assert.Equal(t, float64(456), calculateTOUValue(
		timeOfUse{
			Name:         "test",
			Description:  "test description",
			DefaultValue: 123,
			TimeWindows: []timeWindow{{
				Value:         456,
				Start:         "11:59",
				End:           "12:01",
				startDuration: time.Duration(11*time.Hour + 59*time.Minute),
				endDuration:   time.Duration(12*time.Hour + 1*time.Minute),
			}},
		},
		time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC),
	), "should return timeWindow value")
	assert.Equal(t, float64(789), calculateTOUValue(timeOfUse{
		Name:         "test",
		Description:  "test description",
		DefaultValue: 123,
		TimeWindows: []timeWindow{
			{
				Value:         456,
				Start:         "07:00",
				End:           "12:00",
				startDuration: time.Duration(7 * time.Hour),
				endDuration:   time.Duration(12 * time.Hour),
			},
			{
				Value:         789,
				Start:         "12:00",
				End:           "17:00",
				startDuration: time.Duration(12 * time.Hour),
				endDuration:   time.Duration(17 * time.Hour),
			},
		}},
		time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC),
	), "boundary check, should return timeWindow value")

}

func TestIsWIthinTimeWindow(t *testing.T) {
	assert.Equal(t, true, isWithinTimeWindow(
		timeWindow{
			Value:         0,
			Start:         "23:57",
			End:           "23:59",
			startDuration: time.Duration(23*time.Hour + 57*time.Minute),
			endDuration:   time.Duration(23*time.Hour + 59*time.Minute),
		},
		time.Date(2023, 12, 9, 23, 58, 0, 0, time.UTC),
	), "should be within time window")
	assert.Equal(t, true, isWithinTimeWindow(
		timeWindow{
			Value:         0,
			Start:         "23:57",
			End:           "23:59",
			startDuration: time.Duration(23*time.Hour + 57*time.Minute),
			endDuration:   time.Duration(23*time.Hour + 59*time.Minute),
		},
		time.Date(2023, 12, 9, 23, 57, 0, 0, time.UTC),
	), "should be within time window, boundary check start")
	assert.Equal(t, false, isWithinTimeWindow(
		timeWindow{
			Value:         0,
			Start:         "23:57",
			End:           "23:59",
			startDuration: time.Duration(23*time.Hour + 57*time.Minute),
			endDuration:   time.Duration(23*time.Hour + 59*time.Minute),
		},
		time.Date(2023, 12, 9, 23, 59, 0, 0, time.UTC),
	), "should not be within time window, boundary check end")

	chatham, err := time.LoadLocation("Pacific/Chatham")
	if err != nil {
		t.Error("error loading timezone", "err", err)
	}
	assert.Equal(t, true, isWithinTimeWindow(
		timeWindow{
			Value:         0,
			Start:         "23:57",
			End:           "23:59",
			startDuration: time.Duration(23*time.Hour + 57*time.Minute),
			endDuration:   time.Duration(23*time.Hour + 59*time.Minute),
		},
		time.Date(2023, 12, 9, 23, 58, 0, 0, chatham),
	), "Check within time window in Chatham timezone")
	assert.Equal(t, true, isWithinTimeWindow(
		timeWindow{
			Value:         0,
			Start:         "00:00",
			End:           "00:01",
			startDuration: time.Duration(0),
			endDuration:   time.Duration(1 * time.Minute),
		},
		time.Date(2023, 12, 9, 0, 0, 0, 1, chatham),
	), "Check within time window in Chatham timezone")
}
