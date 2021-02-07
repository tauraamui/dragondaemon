package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/franela/goblin"
	"github.com/tauraamui/dragondaemon/config/schedule"
)

func TestSchedule(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("Configuration schedule time checking", func() {
		mockValidConfigWithSchedule := []byte(`{
				"cameras": [
					{
						"schedule": {
							"tuesday": {
								"off": "09:00:00"
							},
							"saturday": {
								"on": "19:00:00"
							}
						}
					}
				]
			}`)

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

		g.It("Should return off if given time after off time on Tuesday", func() {
			// back date today to Tuesday 2nd Feb 2021
			schedule.TODAY = time.Date(2021, 02, 2, 0, 0, 0, 0, time.UTC)

			err := cfg.Load()
			g.Assert(err).IsNil()

			camera := cfg.Cameras[0]
			g.Assert(camera).IsNotNil()
			g.Assert(camera.Schedule).IsNotNil()

			currentTime := time.Date(2021, 02, 2, 9, 10, 0, 0, time.UTC)
			g.Assert(camera.Schedule.IsOn(schedule.Time(currentTime))).IsFalse()
		})

		g.It("Should return off if given time on Wednesday is after off time on Tuesday", func() {
			schedule.TODAY = time.Date(2021, 02, 3, 0, 0, 0, 0, time.UTC)

			err := cfg.Load()
			g.Assert(err).IsNil()

			camera := cfg.Cameras[0]
			g.Assert(camera).IsNotNil()
			g.Assert(camera.Schedule).IsNotNil()

			currentTime := time.Date(2021, 02, 3, 7, 0, 0, 0, time.UTC)
			g.Assert(camera.Schedule.IsOn(schedule.Time(currentTime))).IsFalse()
		})
	})
}
