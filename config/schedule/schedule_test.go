package schedule

import (
	"testing"
	"time"

	"github.com/franela/goblin"
)

func TestSchedule(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("isTimeOnOrOff", func() {
		g.It("Should return on if current time after nil", func() {
			empty, onOrOff := isTimeOnOrOff(Time(time.Now()), nil)
			g.Assert(empty).IsTrue()
			g.Assert(onOrOff).IsTrue()
		})

		g.It("Should return on if current time before off", func() {
			currentTime := time.Date(2021, 9, 13, 11, 0, 0, 0, time.UTC)
			offTime := Time(time.Date(2021, 9, 13, 12, 0, 0, 0, time.UTC))

			empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
				Off: &offTime,
			})

			g.Assert(empty).IsFalse()
			g.Assert(onOrOff).IsTrue()
		})

		g.It("Should return off if current time after off", func() {
			currentTime := time.Date(2021, 9, 13, 13, 0, 0, 0, time.UTC)
			offTime := Time(time.Date(2021, 9, 13, 9, 0, 0, 0, time.UTC))

			empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
				Off: &offTime,
			})

			g.Assert(empty).IsFalse()
			g.Assert(onOrOff).IsFalse()
		})

		g.It("Should return on if current time after on", func() {
			currentTime := time.Date(2021, 9, 13, 13, 0, 0, 0, time.UTC)
			onTime := Time(time.Date(2021, 9, 13, 11, 0, 0, 0, time.UTC))

			empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
				On: &onTime,
			})

			g.Assert(empty).IsFalse()
			g.Assert(onOrOff).IsTrue()
		})

		g.It("Should return off if current time before on after off", func() {
			currentTime := time.Date(2021, 9, 13, 13, 0, 0, 0, time.UTC)
			offTime := Time(time.Date(2021, 9, 13, 10, 0, 0, 0, time.UTC))
			onTime := Time(time.Date(2021, 9, 13, 14, 0, 0, 0, time.UTC))

			empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
				On:  &onTime,
				Off: &offTime,
			})

			g.Assert(empty).IsFalse()
			g.Assert(onOrOff).IsFalse()
		})

		g.It("Should return on if current time after on after off", func() {
			currentTime := time.Date(2021, 9, 13, 18, 0, 0, 0, time.UTC)
			offTime := Time(time.Date(2021, 9, 13, 13, 0, 0, 0, time.UTC))
			onTime := Time(time.Date(2021, 9, 13, 14, 0, 0, 0, time.UTC))

			empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
				On:  &onTime,
				Off: &offTime,
			})

			g.Assert(empty).IsFalse()
			g.Assert(onOrOff).IsTrue()
		})
	})

	// g.Describe("Configuration schedule time from 10am off to 5pm on Sunday", func() {
	// 	mockSchedule := []byte(`{
	// 		"sunday": {
	// 			"off": "10:00:00",
	// 			"on": "17:00:00"
	// 		}
	// 	}`)

	// 	g.It("Should return on if given time on Sunday before same day off time", func() {
	// 		TODAY = time.Date(2021, 2, 7, 0, 0, 0, 0, time.UTC)

	// 		testSchedule := Schedule{}
	// 		err := json.Unmarshal(mockSchedule, &testSchedule)
	// 		g.Assert(err).IsNil()

	// 		currentTime := time.Date(2021, 2, 7, 9, 0, 0, 0, time.UTC)
	// 		g.Assert(testSchedule.IsOn(Time(currentTime))).IsTrue()
	// 	})

	// 	g.It("Should return off if given time on Sunday after same day off time", func() {
	// 		TODAY = time.Date(2021, 2, 7, 0, 0, 0, 0, time.UTC)

	// 		testSchedule := Schedule{}
	// 		err := json.Unmarshal(mockSchedule, &testSchedule)
	// 		g.Assert(err).IsNil()

	// 		currentTime := time.Date(2021, 2, 7, 11, 0, 0, 0, time.UTC)
	// 		g.Assert(testSchedule.IsOn(Time(currentTime))).IsFalse()
	// 	})

	// 	g.It("Should return on if given time on Sunday after same day on time", func() {
	// 		TODAY = time.Date(2021, 2, 7, 0, 0, 0, 0, time.UTC)

	// 		testSchedule := Schedule{}
	// 		err := json.Unmarshal(mockSchedule, &testSchedule)
	// 		g.Assert(err).IsNil()

	// 		currentTime := time.Date(2021, 2, 7, 17, 20, 0, 0, time.UTC)
	// 		g.Assert(testSchedule.IsOn(Time(currentTime))).IsTrue()
	// 	})
	// })

	// g.Describe("Configuration schedule time from Tuesday 9am off to Saturday 7pm", func() {
	// 	mockValidConfigWithSchedule := []byte(`{
	// 		"tuesday": {
	// 			"off": "09:00:00"
	// 		},
	// 		"saturday": {
	// 			"on": "19:00:00"
	// 		}
	// 	}`)

	// 	g.It("Should return on if given time on Monday before off time Tuesday", func() {
	// 		// back date today to Monday 1nd Feb 2021
	// 		TODAY = time.Date(2021, 2, 2, 0, 0, 0, 0, time.UTC)

	// 		testSchedule := Schedule{}
	// 		err := json.Unmarshal(mockValidConfigWithSchedule, &testSchedule)
	// 		g.Assert(err).IsNil()

	// 		currentTime := time.Date(2021, 2, 1, 23, 50, 0, 0, time.UTC)
	// 		g.Assert(testSchedule.IsOn(Time(currentTime))).IsTrue()
	// 	})

	// 	g.It("Should return on if given same day time before off time Tuesday", func() {
	// 		// back date today to Tuesday 2nd Feb 2021
	// 		TODAY = time.Date(2021, 2, 2, 0, 0, 0, 0, time.UTC)

	// 		testSchedule := Schedule{}
	// 		err := json.Unmarshal(mockValidConfigWithSchedule, &testSchedule)
	// 		g.Assert(err).IsNil()

	// 		currentTime := time.Date(2021, 2, 2, 8, 50, 0, 0, time.UTC)
	// 		g.Assert(testSchedule.IsOn(Time(currentTime))).IsTrue()
	// 	})

	// 	g.It("Should return off if given same day time after off time on Tuesday", func() {
	// 		// back date today to Tuesday 2nd Feb 2021
	// 		TODAY = time.Date(2021, 2, 2, 0, 0, 0, 0, time.UTC)

	// 		testSchedule := Schedule{}
	// 		err := json.Unmarshal(mockValidConfigWithSchedule, &testSchedule)
	// 		g.Assert(err).IsNil()

	// 		currentTime := time.Date(2021, 2, 2, 9, 10, 0, 0, time.UTC)
	// 		g.Assert(testSchedule.IsOn(Time(currentTime))).IsFalse()
	// 	})

	// 	g.It("Should return off if given time on Wednesday is after off time on Tuesday", func() {
	// 		TODAY = time.Date(2021, 2, 3, 0, 0, 0, 0, time.UTC)

	// 		testSchedule := Schedule{}
	// 		err := json.Unmarshal(mockValidConfigWithSchedule, &testSchedule)
	// 		g.Assert(err).IsNil()

	// 		currentTime := time.Date(2021, 2, 3, 7, 0, 0, 0, time.UTC)
	// 		g.Assert(testSchedule.IsOn(Time(currentTime))).IsFalse()
	// 	})

	// 	g.It("Should return off if given time on Thursday is after off time on Tuesday", func() {
	// 		TODAY = time.Date(2021, 2, 4, 0, 0, 0, 0, time.UTC)

	// 		testSchedule := Schedule{}
	// 		err := json.Unmarshal(mockValidConfigWithSchedule, &testSchedule)
	// 		g.Assert(err).IsNil()

	// 		currentTime := time.Date(2021, 2, 4, 7, 0, 0, 0, time.UTC)
	// 		g.Assert(testSchedule.IsOn(Time(currentTime))).IsFalse()
	// 	})

	// 	g.It("Should return off if given time on Friday is after off time on Tuesday", func() {
	// 		TODAY = time.Date(2021, 2, 5, 0, 0, 0, 0, time.UTC)

	// 		testSchedule := Schedule{}
	// 		err := json.Unmarshal(mockValidConfigWithSchedule, &testSchedule)
	// 		g.Assert(err).IsNil()

	// 		currentTime := time.Date(2021, 2, 5, 7, 0, 0, 0, time.UTC)
	// 		g.Assert(testSchedule.IsOn(Time(currentTime))).IsFalse()
	// 	})

	// 	g.It("Should return off if given time on Saturday is after off time on Tuesday and before same day on time", func() {
	// 		TODAY = time.Date(2021, 2, 6, 0, 0, 0, 0, time.UTC)

	// 		testSchedule := Schedule{}
	// 		err := json.Unmarshal(mockValidConfigWithSchedule, &testSchedule)
	// 		g.Assert(err).IsNil()

	// 		currentTime := time.Date(2021, 2, 6, 7, 0, 0, 0, time.UTC)
	// 		g.Assert(testSchedule.IsOn(Time(currentTime))).IsFalse()
	// 	})

	// 	g.It("Should return on if given time on Saturday is after off time on Tuesday and after same day on time", func() {
	// 		TODAY = time.Date(2021, 2, 6, 0, 0, 0, 0, time.UTC)

	// 		testSchedule := Schedule{}
	// 		err := json.Unmarshal(mockValidConfigWithSchedule, &testSchedule)
	// 		g.Assert(err).IsNil()

	// 		currentTime := time.Date(2021, 2, 6, 19, 20, 0, 0, time.UTC)
	// 		g.Assert(testSchedule.IsOn(Time(currentTime))).IsTrue()
	// 	})

	// 	g.It("Should return on if given time on Sunday is after on time on Saturday", func() {
	// 		TODAY = time.Date(2021, 2, 7, 0, 0, 0, 0, time.UTC)

	// 		testSchedule := Schedule{}
	// 		err := json.Unmarshal(mockValidConfigWithSchedule, &testSchedule)
	// 		g.Assert(err).IsNil()

	// 		currentTime := time.Date(2021, 2, 7, 7, 0, 0, 0, time.UTC)
	// 		g.Assert(testSchedule.IsOn(Time(currentTime))).IsTrue()
	// 	})
	// })
}
