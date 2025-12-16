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
						TimeWindows: []timeWindow{
							{
								Value:     0.2423,
								Start:     "11:00",
								End:       "21:00",
								startHour: 11,
								endHour:   21,
							},
							{
								Value:     0.2423,
								Start:     "07:00",
								End:       "11:00",
								startHour: 7,
								endHour:   11,
								Days:      []int{2, 3},
							},
						},
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

func TestParseWindowTimes(t *testing.T) {
	testCases := map[string]struct {
		input     string
		expectedH int
		expectedM int
		err       error
	}{
		"standard": {
			input:     "13:00",
			expectedH: 13,
			expectedM: 0,
			err:       nil,
		},
		"multiple units": {
			input:     "15:30",
			expectedH: 15,
			expectedM: 30,
			err:       nil,
		},
		"unsupported second resolution": {
			input:     "15:30:00",
			expectedH: 0,
			expectedM: 0,
			err:       errors.New(`Invalid time format. Must be hh:mm. Got: "15:30:00"`),
		},
		"duration": {
			input:     "19h13m",
			expectedH: 0,
			expectedM: 0,
			err:       errors.New(`Invalid time format. Must be hh:mm. Got: "19h13m"`),
		},
		"empty string": {
			input:     "",
			expectedH: 0,
			expectedM: 0,
			err:       errors.New(`Invalid time format. Must be hh:mm. Got: ""`),
		},
	}

	for name, tc := range testCases {
		actualH, actualM, err := parseWindowTimes(tc.input)
		assert.Equal(t, tc.expectedH, actualH, name)
		assert.Equal(t, tc.expectedM, actualM, name)
		assert.Equal(t, tc.err, err, name)
	}
}
