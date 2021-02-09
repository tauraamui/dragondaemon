package config

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/franela/goblin"
	"github.com/tauraamui/dragondaemon/config/schedule"
	"gopkg.in/dealancer/validate.v2"
)

func TestConfig(t *testing.T) {
	g := goblin.Goblin(t)
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

}
