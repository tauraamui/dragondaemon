package config

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	. "github.com/franela/goblin"
	"github.com/tauraamui/dragondaemon/config/schedule"
	"gopkg.in/dealancer/validate.v2"
)

func Test(t *testing.T) {
	g := Goblin(t)
	g.Describe("Loading configuration from file", func() {

		mockValidConfigContent := []byte(`{
				"debug": true,
				"cameras": [
					{
						"title": "Test Cam 1",
						"address": "camera-network-addr",
						"fps": 1,
						"seconds_per_clip": 2,
						"disabled": false,
						"schedule": {
							"monday": {
								"on": "08:00:00",
								"off": "19:00:00"
							}
						}
					}
				]
			}`)

		mockInvalidJSONConfigContent := []byte(`{
			"debug" true,
		}`)

		mockValidationErroringConfigContent := []byte(`{
			"cameras": [
				{
					"title": "Test Cam 2"
				}
			]
		}`)

		g.It("Should pass the expected ENV value for config location into file reader", func() {
			// set the ENV var to known value
			os.Setenv("DRAGON_DAEMON_CONFIG", "test-config-path")

			cfg := values{
				r: func(path string) ([]byte, error) {
					g.Assert(path).Equal("test-config-path")
					return []byte{}, nil
				},
				um: json.Unmarshal,
				v:  validate.Validate,
			}

			cfg.Load()
		})

		g.It("Should load values from config file into struct", func() {
			cfg := values{
				r: func(string) ([]byte, error) {
					return mockValidConfigContent, nil
				},
				um: json.Unmarshal,
				v:  validate.Validate,
			}

			err := cfg.Load()
			g.Assert(err).IsNil()
			g.Assert(cfg.Debug).IsTrue()
			g.Assert(cfg.Cameras).Equal([]Camera{
				{
					Title:          "Test Cam 1",
					Address:        "camera-network-addr",
					FPS:            1,
					SecondsPerClip: 2,
					Disabled:       false,
					Schedule: schedule.Schedule{
						Monday: schedule.OnOffTimes{
							On: func() *schedule.Time {
								t, _ := schedule.ParseTime("08:00:00")
								return &t
							}(),
							Off: func() *schedule.Time {
								t, _ := schedule.ParseTime("19:00:00")
								return &t
							}(),
						},
					},
				},
			})
		})

		g.It("Should return error if unable to read configuration", func() {
			cfg := values{
				r: func(string) ([]byte, error) {
					return nil, errors.New("read failure")
				},
			}

			err := cfg.Load()
			g.Assert(err).IsNotNil()
			g.Assert(err.Error()).Equal("Unable to read from path test-config-path: read failure")
		})

		g.It("Should return error if unable to unmarshal JSON into configuration struct", func() {
			cfg := values{
				r: func(string) ([]byte, error) {
					return mockInvalidJSONConfigContent, nil
				},
				um: json.Unmarshal,
			}

			err := cfg.Load()
			g.Assert(err).IsNotNil()
			g.Assert(err.Error()).Equal("Parsing configuration file error: invalid character 't' after object key")
		})

		g.It("Should return error if configuration unable to pass validation", func() {
			cfg := values{
				r: func(string) ([]byte, error) {
					return mockValidationErroringConfigContent, nil
				},
				um: json.Unmarshal,
				v:  validate.Validate,
			}

			err := cfg.Load()
			g.Assert(err).IsNotNil()
			g.Assert(err.Error()).Equal(
				"Unable to validate configuration: Validation error in field \"FPS\" of type \"int\" using validator \"gte=1\"",
			)
		})
	})

	g.Describe("Configuration schedule time checking", func() {
		mockValidConfigWithSchedule := []byte(`{
				"cameras": [
					{
						"schedule": {
							"monday": {
								"on": "09:30:00"
							}
						}
					}
				]
			}`)

		g.It("Camera is on given time after on time on Monday", func() {
			cfg := values{
				r: func(string) ([]byte, error) {
					return mockValidConfigWithSchedule, nil
				},
				um: json.Unmarshal,
				// disable validation to allow for smaller mock config
				v: func(interface{}) error {
					return nil
				},
			}

			// back date today to Monday 1st Feb 2021
			schedule.TODAY = time.Date(2021, 02, 1, 0, 0, 0, 0, time.UTC)

			err := cfg.Load()
			g.Assert(err).IsNil()

			camera := cfg.Cameras[0]
			g.Assert(camera).IsNotNil()
			g.Assert(camera.Schedule).IsNotNil()

			currentTime := time.Date(2021, 02, 1, 13, 0, 0, 0, time.Now().Location())
			g.Assert(camera.Schedule.IsOn(schedule.Time(currentTime))).IsTrue()
		})
	})
}
