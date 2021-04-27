package schedule

import (
	"testing"
	"time"

	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestSchedule(t *testing.T) {
	g := goblin.Goblin(t)

	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("isTimeOnOrOff", func() {
		g.It("Returns on if current time after nil", func() {
			empty, onOrOff := isTimeOnOrOff(Time(time.Now()), nil)
			Expect(empty).To(BeTrue())
			Expect(onOrOff).To(BeTrue())
		})

		g.It("Returns on if current time before off", func() {
			currentTime := time.Date(2021, 9, 13, 11, 0, 0, 0, time.UTC)
			offTime := Time(time.Date(2021, 9, 13, 12, 0, 0, 0, time.UTC))

			empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
				Off: &offTime,
			})

			Expect(empty).To(BeTrue())
			Expect(onOrOff).To(BeTrue())
		})

		g.It("Returns off if current time after off", func() {
			currentTime := time.Date(2021, 9, 13, 13, 0, 0, 0, time.UTC)
			offTime := Time(time.Date(2021, 9, 13, 9, 0, 0, 0, time.UTC))

			empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
				Off: &offTime,
			})

			Expect(empty).To(BeFalse())
			Expect(onOrOff).To(BeFalse())
		})

		g.It("Returns on if current time after on", func() {
			currentTime := time.Date(2021, 9, 13, 13, 0, 0, 0, time.UTC)
			onTime := Time(time.Date(2021, 9, 13, 11, 0, 0, 0, time.UTC))

			empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
				On: &onTime,
			})

			Expect(empty).To(BeFalse())
			Expect(onOrOff).To(BeTrue())
		})

		g.It("Returns off if current time after off", func() {
			currentTime := time.Date(2021, 9, 13, 13, 0, 0, 0, time.UTC)
			offTime := Time(time.Date(2021, 9, 13, 11, 0, 0, 0, time.UTC))

			empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
				Off: &offTime,
			})

			Expect(empty).To(BeFalse())
			Expect(onOrOff).To(BeFalse())
		})

		g.It("Returns off if current time before on after off", func() {
			currentTime := time.Date(2021, 9, 13, 13, 0, 0, 0, time.UTC)
			offTime := Time(time.Date(2021, 9, 13, 10, 0, 0, 0, time.UTC))
			onTime := Time(time.Date(2021, 9, 13, 14, 0, 0, 0, time.UTC))

			empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
				On:  &onTime,
				Off: &offTime,
			})

			Expect(empty).To(BeFalse())
			Expect(onOrOff).To(BeFalse())
		})

		g.It("Returns on if current time after on before off", func() {
			currentTime := time.Date(2021, 9, 13, 14, 0, 0, 0, time.UTC)
			offTime := Time(time.Date(2021, 9, 13, 17, 0, 0, 0, time.UTC))
			onTime := Time(time.Date(2021, 9, 13, 13, 0, 0, 0, time.UTC))

			empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
				On:  &onTime,
				Off: &offTime,
			})

			Expect(empty).To(BeFalse())
			Expect(onOrOff).To(BeTrue())
		})

		g.It("Returns on if current time after on after off", func() {
			currentTime := time.Date(2021, 9, 13, 18, 0, 0, 0, time.UTC)
			offTime := Time(time.Date(2021, 9, 13, 13, 0, 0, 0, time.UTC))
			onTime := Time(time.Date(2021, 9, 13, 14, 0, 0, 0, time.UTC))

			empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
				On:  &onTime,
				Off: &offTime,
			})

			Expect(empty).To(BeFalse())
			Expect(onOrOff).To(BeTrue())
		})
	})
}
