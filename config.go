package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

// type config struct {
// 	LocalizedTimezones []string `yaml:"localized_timezones"`
// }

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
	Value         float64 `yaml:"value"`
	Start         string  `yaml:"start"`
	End           string  `yaml:"end"`
	startDuration time.Duration
	endDuration   time.Duration
}

var liveConfig = config{}

func configInit() {
	f := "config.yaml"
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
				slog.Info("Watcher error", "err", err, "ok", ok)
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
		for j, tw := range tou.TimeWindows {
			c.TimeOfUse[i].TimeWindows[j].startDuration, err = calculateDuration(tw.Start)
			if err != nil {
				slog.Error("Error parsing time window start", "err", err, "time_of_use", tou.Name, "time_window", tw)
				return config{}, err
			}

			c.TimeOfUse[i].TimeWindows[j].endDuration, err = calculateDuration(tw.End)
			if err != nil {
				slog.Error("Error parsing time window end", "err", err, "time_of_use", tou.Name, "time_window", tw)
				return config{}, err
			}
		}
	}

	return c, nil
}

func calculateDuration(t string) (time.Duration, error) {
	d, err := time.ParseDuration(t)
	if err != nil {
		return 0, err
	}
	return d, nil
}
