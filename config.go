package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

type config struct {
	LocalizedTimezones []string    `yaml:"localized_timezones"`
	TimeOfUse          []timeOfUse `yaml:"time_of_use,omitempty"`
}

type timeOfUse struct {
	Name         string            `yaml:"name"`
	Description  string            `yaml:"description"`
	Timezone     string            `yaml:"timezone,omitempty"`
	Labels       map[string]string `yaml:"labels,omitempty"`
	DefaultValue float64           `yaml:"default_value"`
	TimeWindows  []timeWindow      `yaml:"time_windows"`
}

type timeWindow struct {
	Value       float64           `yaml:"value"`
	Start       string            `yaml:"start"`
	End         string            `yaml:"end"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Days        []int             `yaml:"days,omitempty"`
	startHour   int
	startMinute int
	endHour     int
	endMinute   int
}

var liveConfig = config{}

func configInit() {
	f := "./config.yaml"
	if os.Getenv("CONFIG_FILE") != "" {
		f = os.Getenv("CONFIG_FILE")
	}
	c, err := loadConfig(f)
	if err != nil {
		slog.Error("Error loading config", "err", err)
		os.Exit(1)
	}
	liveConfig = c
	go configWatcher(f)
}

func configWatcher(filepath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("Error creating new watcher", "err", err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				slog.Debug("Watcher event", "event", event, "ok", ok)
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					c, err := loadConfig(filepath)
					if err != nil {
						slog.Error("Error loading config", "err", err)
						continue
					}
					liveConfig = c
				}
			case err, ok := <-watcher.Errors:
				slog.Debug("Watcher error", "err", err, "ok", ok)
				if !ok {
					return
				}
				slog.Error("Error watching config", "err", err)
			}
		}
	}()

	err = watcher.Add(filepath)
	if err != nil {
		slog.Error("Error adding filepath to watcher", "err", err)
	}
	slog.Debug("Added config watcher", "filepath", filepath)

	// Block main goroutine
	<-make(chan struct{})
}

func loadConfig(filepath string) (config, error) {
	slog.Info("Loading config", "filepath", filepath)
	f, err := os.ReadFile(filepath)
	if err != nil {
		return config{}, err
	}

	c := config{}
	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return config{}, err
	}

	for _, loc := range c.LocalizedTimezones {
		_, err := time.LoadLocation(loc)
		if err != nil {
			return config{}, err
		}
	}

	for i, tou := range c.TimeOfUse {
		_, err := time.LoadLocation(tou.Timezone)
		if err != nil {
			slog.Error("Error parsing timezone", "err", err, "time_of_use", tou.Name, "timezone", tou.Timezone)
			return config{}, err
		}

		for j, tw := range tou.TimeWindows {
			slog.Debug("Parsing time window", "time_of_use", tou.Name, "time_window", tw)
			c.TimeOfUse[i].TimeWindows[j].startHour, c.TimeOfUse[i].TimeWindows[j].startMinute, err = parseWindowTimes(tw.Start)
			if err != nil {
				slog.Error("Error parsing time window start", "err", err, "time_of_use", tou.Name, "time_window", tw)
				return config{}, err
			}

			c.TimeOfUse[i].TimeWindows[j].endHour, c.TimeOfUse[i].TimeWindows[j].endMinute, err = parseWindowTimes(tw.End)
			if err != nil {
				slog.Error("Error parsing time window end", "err", err, "time_of_use", tou.Name, "time_window", tw)
				return config{}, err
			}
		}
	}

	return c, nil
}

func parseWindowTimes(t string) (int, int, error) {
	// Split string by :
	parts := strings.Split(t, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf(`Invalid time format. Must be hh:mm. Got: "%s"`, t)
	}
	// Convert both sides to int
	h, err := strconv.Atoi(parts[0])
	if err != nil || h < 0 || h > 23 {
		return 0, 0, fmt.Errorf(
			`Error when parsing hh. Invalid hour format. Must be a number, and 0-23. Got: "%s" err: %s`,
			parts[0],
			err.Error(),
		)
	}

	m, err := strconv.Atoi(parts[1])
	if err != nil || m < 0 || m > 59 {
		return 0, 0, errors.New(fmt.Sprintf(
			`Error when parsing mm. Invalid minute format. Must be a number, and 0-59. Got: "%s" err: %s`,
			parts[1],
			err.Error(),
		))
	}

	return h, m, nil
}
