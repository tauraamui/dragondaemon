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
							"monday": {
								"on": "09:30:00"
							},
							"tuesday": {
								"off": "17:00:00"
							},
							"wednesday": {
								"off": "10:30:00",
								"on": "15:00:00"
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

		g.It("Camera is on if given time after on time on Monday", func() {
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

		g.It("Camera is off if given time after off time on Tuesday", func() {
			// back date today to Tuesday 2nd Feb 2021
			schedule.TODAY = time.Date(2021, 02, 2, 0, 0, 0, 0, time.UTC)

			err := cfg.Load()
			g.Assert(err).IsNil()

			camera := cfg.Cameras[0]
			g.Assert(camera).IsNotNil()
			g.Assert(camera.Schedule).IsNotNil()

			currentTime := time.Date(2021, 02, 2, 17, 10, 0, 0, time.Now().Location())
			g.Assert(camera.Schedule.IsOn(schedule.Time(currentTime))).IsFalse()
		})

		g.It("Camera is on if given time is after on time which is later than off time", func() {
			// back date today to Wednesday 3rd Feb 2021
			schedule.TODAY = time.Date(2021, 02, 3, 0, 0, 0, 0, time.UTC)

			err := cfg.Load()
			g.Assert(err).IsNil()

			camera := cfg.Cameras[0]
			g.Assert(camera).IsNotNil()
			g.Assert(camera.Schedule).IsNotNil()

			currentTime := time.Date(2021, 02, 3, 15, 10, 0, 0, time.Now().Location())
			g.Assert(camera.Schedule.IsOn(schedule.Time(currentTime))).IsTrue()
		})
	})
}
