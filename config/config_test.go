package config

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	. "github.com/franela/goblin"
	"gopkg.in/dealancer/validate.v2"
)

func Test(t *testing.T) {
	g := Goblin(t)
	g.Describe("Configuration loading", func() {

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
								"on": "08:00",
								"off": "19:00"
							}
						}
					}
				]
			}`)

		g.It("Should pass the expected ENV value for config location into file reader", func() {
			// set the ENV var to known value
			os.Setenv("DRAGON_DAEMON_CONFIG", "test-config-path")

			cfg := Config{
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
			cfg := Config{
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
					Schedule: Schedule{
						Monday: OnOffTimes{
							On:  "08:00",
							Off: "19:00",
						},
					},
				},
			})
		})

		g.It("Should return error if unable to read configuration", func() {
			cfg := Config{
				r: func(string) ([]byte, error) {
					return nil, errors.New("read failure")
				},
			}

			err := cfg.Load()
			g.Assert(err).IsNotNil()
			g.Assert(err.Error()).Equal("read failure")
		})
	})
}
