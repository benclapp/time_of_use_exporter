package main

import (
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Localized timezones
	minuteLocalized     = prometheus.NewDesc("tou_exporter_localized_minute", "Minute of the hour from 0-59 in a specific timezone", []string{"tz"}, nil)
	hourLocalized       = prometheus.NewDesc("tou_exporter_localized_hour", "Hour of the day from 0-23 in a specific timezone", []string{"tz"}, nil)
	dayOfWeekLocalized  = prometheus.NewDesc("tou_exporter_localized_day_of_week", "Day of the week from 0-6 in a specific timezone. 0 is Sunday.", []string{"tz", "day"}, nil)
	dayOfMonthLocalized = prometheus.NewDesc("tou_exporter_localized_day_of_month", "Day of the month from 1-31 in a specific timezone", []string{"tz"}, nil)
	monthLocalized      = prometheus.NewDesc("tou_exporter_localized_month", "Month of the year from 1-12 in a specific timezone", []string{"tz", "month"}, nil)
)

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	describeLocalizedTimezones(ch)
	describeTOUMetrics(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	utc, err := time.LoadLocation("UTC")
	if err != nil {
		slog.Error("error loading UTC timezone", "err", err)
		return
	}
	t := time.Now().In(utc)

	collectLocalizedTimezones(ch, t)
	collectTOUMetrics(ch, t)
}

func describeLocalizedTimezones(ch chan<- *prometheus.Desc) {
	ch <- minuteLocalized
	ch <- hourLocalized
	ch <- dayOfWeekLocalized
	ch <- dayOfMonthLocalized
	ch <- monthLocalized
}

func collectLocalizedTimezones(ch chan<- prometheus.Metric, utcNow time.Time) {
	for _, tz := range liveConfig.LocalizedTimezones {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			slog.Error("error loading timezone", "tz", tz, "err", err)
			continue
		}
		ch <- prometheus.MustNewConstMetric(minuteLocalized, prometheus.GaugeValue, float64(utcNow.In(loc).Minute()), tz)
		ch <- prometheus.MustNewConstMetric(hourLocalized, prometheus.GaugeValue, float64(utcNow.In(loc).Hour()), tz)
		ch <- prometheus.MustNewConstMetric(dayOfWeekLocalized, prometheus.GaugeValue, float64(utcNow.In(loc).Weekday()), tz, utcNow.In(loc).Weekday().String())
		ch <- prometheus.MustNewConstMetric(dayOfMonthLocalized, prometheus.GaugeValue, float64(utcNow.In(loc).Day()), tz)
		ch <- prometheus.MustNewConstMetric(monthLocalized, prometheus.GaugeValue, float64(utcNow.In(loc).Month()), tz, utcNow.In(loc).Month().String())
	}
}

func describeTOUMetrics(ch chan<- *prometheus.Desc) {
	for _, tou := range liveConfig.TimeOfUse {
		ch <- describeTOUMetric(tou)
	}
}

func collectTOUMetrics(ch chan<- prometheus.Metric, utcNow time.Time) {
	for _, tou := range liveConfig.TimeOfUse {
		loc, err := time.LoadLocation(tou.Timezone)
		if err != nil {
			slog.Error("error loading timezone. This should never error as TZ are validated on config load", "err", err, "timezon", tou.Timezone)
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			describeTOUMetric(tou),
			prometheus.GaugeValue,
			calculateTOUValue(tou, utcNow.In(loc)),
		)
	}
}
