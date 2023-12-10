package main

import (
	"errors"
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
	c, err := loadConfig("config_test.yaml")
	if assert.NoError(t, err) {
		require.IsType(t, config{}, c)
		require.EqualValues(
			t, config{
				LocalizedTimezones: []string{
					"Pacific/Auckland",
				},
				TimeOfUse: []timeOfUse{
					{
						Name:         "electricity_price",
						Description:  "Electricity price",
						Timezone:     "Pacific/Auckland",
						Labels:       map[string]string{"provider": "Power Co"},
						DefaultValue: 0.1106,
						TimeWindows: []timeWindow{{
							Value:         0.2423,
							Start:         "7h",
							End:           "21h",
							startDuration: 7 * time.Hour,
							endDuration:   21 * time.Hour,
						}},
					},
				}}, c,
		)
	}
}

func TestConfigWatcher(t *testing.T) {
	f, err := os.CreateTemp("", "config_test.*.yaml")
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

func TestCalculateDurations(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected time.Duration
		err      error
	}{
		"standard": {
			input:    "13h",
			expected: 13 * time.Hour,
			err:      nil,
		},
		"multiple units": {
			input:    "15h30m5s",
			expected: 15*time.Hour + 30*time.Minute + 5*time.Second,
			err:      nil,
		},
		"time code": {
			input:    "19:00",
			expected: 0,
			err:      errors.New(`time: unknown unit ":" in duration "19:00"`),
		},
		"empty string": {
			input:    "",
			expected: 0,
			err:      errors.New(`time: invalid duration ""`),
		},
	}

	for name, tc := range testCases {
		actual, err := calculateDuration(tc.input)
		assert.Equal(t, tc.expected, actual, name)
		assert.Equal(t, tc.err, err, name)
	}
}
