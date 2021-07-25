package schedule

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Schedule", func() {

	Context("Current time is after nil/unspecified time", func() {
		It("Should return on", func() {
			empty, onOrOff := isTimeOnOrOff(Time(time.Now()), nil)
			Expect(empty).To(BeTrue())
			Expect(onOrOff).To(BeTrue())
		})
	})

	Context("Current time is before off", func() {
		It("Should return on", func() {
			currentTime := time.Date(2021, 9, 13, 11, 0, 0, 0, time.UTC)
			offTime := Time(time.Date(2021, 9, 13, 12, 0, 0, 0, time.UTC))

			empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
				Off: &offTime,
			})

			Expect(empty).To(BeFalse())
			Expect(onOrOff).To(BeTrue())
		})
	})

	Context("Current time is after off", func() {
		It("Should return off", func() {
			currentTime := time.Date(2021, 9, 13, 13, 0, 0, 0, time.UTC)
			offTime := Time(time.Date(2021, 9, 13, 9, 0, 0, 0, time.UTC))

			empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
				Off: &offTime,
			})

			Expect(empty).To(BeFalse())
			Expect(onOrOff).To(BeFalse())
		})
	})

	Context("Current time is after on", func() {
		It("Should return on", func() {
			currentTime := time.Date(2021, 9, 13, 13, 0, 0, 0, time.UTC)
			onTime := Time(time.Date(2021, 9, 13, 11, 0, 0, 0, time.UTC))

			empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
				On: &onTime,
			})

			Expect(empty).To(BeFalse())
			Expect(onOrOff).To(BeTrue())
		})
	})

	Context("Current time is after off", func() {
		It("Should return off", func() {
			currentTime := time.Date(2021, 9, 13, 13, 0, 0, 0, time.UTC)
			offTime := Time(time.Date(2021, 9, 13, 11, 0, 0, 0, time.UTC))

			empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
				Off: &offTime,
			})

			Expect(empty).To(BeFalse())
			Expect(onOrOff).To(BeFalse())
		})
	})

	Context("Current time is before on and after off", func() {
		It("Should return off", func() {
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
	})

	Context("Current time is after on before off", func() {
		It("Should return on", func() {
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
	})

	Context("Current time is after on after off", func() {
		It("Should return on", func() {
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
})
