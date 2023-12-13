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
		"tw labels overrides default": {
			inputTou: timeOfUse{
				Name:        "tw_labels_overrides_default",
				Description: "tw labels overrides default description",
				Labels: map[string]string{
					"foo": "bar", "baz": "qux",
				},
				TimeWindows: []timeWindow{{
					startHour:   11,
					startMinute: 00,
					endHour:     13,
					endMinute:   00,
					Labels: map[string]string{
						"foo": "baz",
					},
				}},
			},
			expectedDesc: prometheus.NewDesc(
				"tw_labels_overrides_default", "tw labels overrides default description",
				nil, map[string]string{"tz": "UTC", "foo": "baz", "baz": "qux"},
			),
		},
	}

	for name, tc := range tcs {
		loc, err := time.LoadLocation(tc.inputTou.Timezone)
		if err != nil {
			t.Fatal("error loading timezone required for test", "err", err)
		}
		v := describeTOUMetric(
			tc.inputTou,
			time.Date(2023, 12, 13, 12, 00, 00, 00, loc),
		)
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
				Value:       456,
				Start:       "11:59",
				End:         "12:01",
				startHour:   11,
				startMinute: 59,
				endHour:     12,
				endMinute:   1,
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
				Value:       456,
				Start:       "11:59",
				End:         "12:01",
				startHour:   11,
				startMinute: 59,
				endHour:     12,
				endMinute:   1,
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
				Value:       456,
				Start:       "07:00",
				End:         "12:00",
				startHour:   7,
				startMinute: 0,
				endHour:     12,
				endMinute:   0,
			},
			{
				Value:       789,
				Start:       "12:00",
				End:         "17:00",
				startHour:   12,
				startMinute: 0,
				endHour:     17,
				endMinute:   0,
			},
		}},
		time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC),
	), "boundary check, should return timeWindow value")

}

func TestIsWIthinTimeWindow(t *testing.T) {
	assert.Equal(t, true, isWithinTimeWindow(
		timeWindow{
			Value:       0,
			Start:       "23:57",
			End:         "23:59",
			startHour:   23,
			startMinute: 57,
			endHour:     23,
			endMinute:   59,
		},
		time.Date(2023, 12, 9, 23, 58, 0, 0, time.UTC),
	), "should be within time window")
	assert.Equal(t, true, isWithinTimeWindow(
		timeWindow{
			Value:       0,
			Start:       "23:57",
			End:         "23:59",
			startHour:   23,
			startMinute: 57,
			endHour:     23,
			endMinute:   59,
		},
		time.Date(2023, 12, 9, 23, 57, 0, 0, time.UTC),
	), "should be within time window, boundary check start")
	assert.Equal(t, false, isWithinTimeWindow(
		timeWindow{
			Value:       0,
			Start:       "23:57",
			End:         "23:59",
			startHour:   23,
			startMinute: 57,
			endHour:     23,
			endMinute:   59,
		},
		time.Date(2023, 12, 9, 23, 59, 0, 0, time.UTC),
	), "should not be within time window, boundary check end")

	chatham, err := time.LoadLocation("Pacific/Chatham")
	if err != nil {
		t.Error("error loading timezone", "err", err)
	}
	assert.Equal(t, true, isWithinTimeWindow(
		timeWindow{
			Value:       0,
			Start:       "23:57",
			End:         "23:59",
			startHour:   23,
			startMinute: 57,
			endHour:     23,
			endMinute:   59,
		},
		time.Date(2023, 12, 9, 23, 58, 0, 0, chatham),
	), "Check within time window in Chatham timezone")
	assert.Equal(t, true, isWithinTimeWindow(
		timeWindow{
			Value:       0,
			Start:       "00:00",
			End:         "00:01",
			startHour:   0,
			startMinute: 0,
			endHour:     0,
			endMinute:   1,
		},
		time.Date(2023, 12, 9, 0, 0, 0, 1, chatham),
	), "Check within time window in Chatham timezone")
}
