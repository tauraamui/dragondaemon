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
	g.Describe("Configuration schedule time from Tuesday 9am off to Saturday 7pm", func() {
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

		g.It("Should return on if given time on Monday before off time Tuesday", func() {
			// back date today to Monday 1nd Feb 2021
			schedule.TODAY = time.Date(2021, 01, 2, 0, 0, 0, 0, time.UTC)

			err := cfg.Load()
			g.Assert(err).IsNil()

			camera := cfg.Cameras[0]
			g.Assert(camera).IsNotNil()
			g.Assert(camera.Schedule).IsNotNil()

			currentTime := time.Date(2021, 02, 1, 23, 50, 0, 0, time.UTC)
			g.Assert(camera.Schedule.IsOn(schedule.Time(currentTime))).IsTrue()
		})

		g.It("Should return on if given same day time before off time Tuesday", func() {
			// back date today to Tuesday 2nd Feb 2021
			schedule.TODAY = time.Date(2021, 02, 2, 0, 0, 0, 0, time.UTC)

			err := cfg.Load()
			g.Assert(err).IsNil()

			camera := cfg.Cameras[0]
			g.Assert(camera).IsNotNil()
			g.Assert(camera.Schedule).IsNotNil()

			currentTime := time.Date(2021, 02, 2, 8, 50, 0, 0, time.UTC)
			g.Assert(camera.Schedule.IsOn(schedule.Time(currentTime))).IsTrue()
		})

		g.It("Should return off if given same day time after off time on Tuesday", func() {
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

		g.It("Should return off if given time on Thursday is after off time on Tuesday", func() {
			schedule.TODAY = time.Date(2021, 02, 4, 0, 0, 0, 0, time.UTC)

			err := cfg.Load()
			g.Assert(err).IsNil()

			camera := cfg.Cameras[0]
			g.Assert(camera).IsNotNil()
			g.Assert(camera.Schedule).IsNotNil()

			currentTime := time.Date(2021, 02, 4, 7, 0, 0, 0, time.UTC)
			g.Assert(camera.Schedule.IsOn(schedule.Time(currentTime))).IsFalse()
		})

		g.It("Should return off if given time on Friday is after off time on Tuesday", func() {
			schedule.TODAY = time.Date(2021, 02, 5, 0, 0, 0, 0, time.UTC)

			err := cfg.Load()
			g.Assert(err).IsNil()

			camera := cfg.Cameras[0]
			g.Assert(camera).IsNotNil()
			g.Assert(camera.Schedule).IsNotNil()

			currentTime := time.Date(2021, 02, 5, 7, 0, 0, 0, time.UTC)
			g.Assert(camera.Schedule.IsOn(schedule.Time(currentTime))).IsFalse()
		})

		g.It("Should return off if given time on Saturday is after off time on Tuesday and before same day on time", func() {
			schedule.TODAY = time.Date(2021, 02, 6, 0, 0, 0, 0, time.UTC)

			err := cfg.Load()
			g.Assert(err).IsNil()

			camera := cfg.Cameras[0]
			g.Assert(camera).IsNotNil()
			g.Assert(camera.Schedule).IsNotNil()

			currentTime := time.Date(2021, 02, 6, 7, 0, 0, 0, time.UTC)
			g.Assert(camera.Schedule.IsOn(schedule.Time(currentTime))).IsFalse()
		})

		g.It("Should return on if given time on Saturday is after off time on Tuesday and after same day on time", func() {
			schedule.TODAY = time.Date(2021, 02, 6, 0, 0, 0, 0, time.UTC)

			err := cfg.Load()
			g.Assert(err).IsNil()

			camera := cfg.Cameras[0]
			g.Assert(camera).IsNotNil()
			g.Assert(camera.Schedule).IsNotNil()

			currentTime := time.Date(2021, 02, 6, 17, 20, 0, 0, time.UTC)
			g.Assert(camera.Schedule.IsOn(schedule.Time(currentTime))).IsTrue()
		})

		g.It("Should return on if given time on Sunday is after on time on Saturday", func() {
			schedule.TODAY = time.Date(2021, 02, 7, 0, 0, 0, 0, time.UTC)

			err := cfg.Load()
			g.Assert(err).IsNil()

			camera := cfg.Cameras[0]
			g.Assert(camera).IsNotNil()
			g.Assert(camera.Schedule).IsNotNil()

			currentTime := time.Date(2021, 02, 7, 7, 0, 0, 0, time.UTC)
			g.Assert(camera.Schedule.IsOn(schedule.Time(currentTime))).IsTrue()
		})
	})
}
