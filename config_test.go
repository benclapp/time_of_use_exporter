package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var testConfig = config{LocalizedTimezones: []string{
	"Pacific/Auckland",
	"Pacific/Chatham",
}}
var testConfigYaml, _ = yaml.Marshal(testConfig)

func TestLoadConfig(t *testing.T) {
	f, err := os.CreateTemp("", "config_test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	os.WriteFile(f.Name(), testConfigYaml, 0644)

	c, err := loadConfig(f.Name())
	if assert.NoError(t, err) {
		require.IsType(t, config{}, c)
		require.EqualValues(t, config{LocalizedTimezones: []string{
			"Pacific/Auckland",
			"Pacific/Chatham",
		}}, c)
	}
}

func TestConfigWatcher(t *testing.T) {
	f, err := os.CreateTemp("", "config_test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	go configWatcher(f.Name())
	time.Sleep(10 * time.Millisecond) // Give the watcher time to start

	assert.Equal(t, config{}, liveConfig)

	err = os.WriteFile(f.Name(), testConfigYaml, 0644)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond) // Give the watcher time to update

	assert.Equal(t, testConfig, liveConfig)
}

func TestConfigInit(t *testing.T) {
	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	os.WriteFile(f.Name(), testConfigYaml, 0644)
	t.Setenv("CONFIG_FILE", f.Name())

	configInit()
	assert.Equal(t, testConfig, liveConfig)
}
