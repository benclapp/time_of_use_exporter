package main

import (
	"fmt"
	"log/slog"
	"os"

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
	Value float64 `yaml:"value"`
	Start string  `yaml:"start"`
	End   string  `yaml:"end"`
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
	fmt.Println("Loading config")
	f, err := os.ReadFile(filepath)
	if err != nil {
		return config{}, err
	}

	c := config{}
	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return config{}, err
	}
	return c, nil
}
