package config

import (
	"encoding/json"
	"os"
	"testing"

	. "github.com/franela/goblin"
	"gopkg.in/dealancer/validate.v2"
)

func Test(t *testing.T) {
	g := Goblin(t)
	g.Describe("Configuration loading", func() {

		mockValidConfigContent := []byte(`{
				"cameras": [
					{
						"title": "Test Cam 1",
						"fps": 1,
						"seconds_per_clip": 2
					}
				]
			}`)

		g.It("Should pass the expected ENV value for config location into file reader", func() {
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
			g.Assert(cfg.Cameras).Equal([]Camera{
				{
					Title:          "Test Cam 1",
					FPS:            1,
					SecondsPerClip: 2,
				},
			})
		})
	})
}
