package schedule

import (
	"testing"
	"time"

	"github.com/matryer/is"
)

var defaultDate timeDate = timeDate{2021, 9, 13}

func timeFromHoursAndMinutes(td timeDate, hour, minute int) Time {
	return Time(time.Date(td.year, time.Month(td.month), td.day, hour, minute, 0, 0, time.UTC))
}

// used in place of unavailable named params langauge feature
type args struct {
	date         timeDate
	hour, minute int
}

type timeDate struct {
	year, month, day int
}

func (td timeDate) empty() bool {
	if td.year == 0 || td.month == 0 || td.day == 0 {
		return true
	}
	return false
}

func testTime(a args) Time {
	date := func() timeDate {
		if a.date.empty() {
			return defaultDate
		}
		return a.date
	}()
	return timeFromHoursAndMinutes(date, a.hour, a.minute)
}

func testTimePtr(a args) *Time {
	tt := testTime(a)
	return &tt
}

type test struct {
	skip              bool
	title             string
	currentTime       Time
	forceEmptyWeekday bool
	onTime            *Time
	offTime           *Time

	isEmpty bool
	isOn    bool
}

func TestDifferentDayScheduleTimesMatchExpectedState(t *testing.T) {
	tests := []test{
		{
			title: "current date+time before off should be on",
			currentTime: testTime(args{
				date: timeDate{
					2021, 3, 15,
				}, hour: 10, minute: 0,
			}),
			offTime: testTimePtr(args{
				date: timeDate{
					2021, 3, 16,
				}, hour: 3, minute: 0,
			}),
			isEmpty: false,
			isOn:    true,
		},
	}

	for _, tt := range tests {
		runTableTestsTest(t, tt)
	}
}

func TestSameDayScheduleTimesMatchExpectedState(t *testing.T) {
	tests := []test{
		{
			title:             "non empty weekday but with no times",
			currentTime:       Time(time.Now()),
			forceEmptyWeekday: true,
			isEmpty:           true,
			isOn:              false,
		},
		{
			title:       "current time is after nil unspecified time should be on",
			currentTime: Time(time.Now()),
			isEmpty:     true,
			isOn:        true,
		},
		{
			title:       "current time is before off should be on",
			currentTime: testTime(args{hour: 11, minute: 0}),
			offTime:     testTimePtr(args{hour: 12, minute: 0}),
			isEmpty:     false,
			isOn:        true,
		},
		{
			title:       "current time is after off should be off",
			currentTime: testTime(args{hour: 13, minute: 0}),
			offTime:     testTimePtr(args{hour: 9, minute: 0}),
			isEmpty:     false,
			isOn:        false,
		},
		{
			title:       "current time is before on should be on",
			currentTime: testTime(args{hour: 13, minute: 0}),
			onTime:      testTimePtr(args{hour: 15, minute: 0}),
			isEmpty:     false,
			isOn:        true,
		},
		{
			title:       "current time is after on should be on",
			currentTime: testTime(args{hour: 13, minute: 0}),
			onTime:      testTimePtr(args{hour: 11, minute: 0}),
			isEmpty:     false,
			isOn:        true,
		},
		{
			title:       "current time is before off after on should be on",
			currentTime: testTime(args{hour: 13, minute: 0}),
			onTime:      testTimePtr(args{hour: 11, minute: 0}),
			offTime:     testTimePtr(args{hour: 15, minute: 0}),
			isEmpty:     false,
			isOn:        true,
		},
		{
			title:       "current time is before on after off should be off",
			currentTime: testTime(args{hour: 13, minute: 0}),
			onTime:      testTimePtr(args{hour: 15, minute: 0}),
			offTime:     testTimePtr(args{hour: 11, minute: 0}),
			isEmpty:     false,
			isOn:        false,
		},
		{
			title:       "current time is after on after off should be on",
			currentTime: testTime(args{hour: 17, minute: 0}),
			onTime:      testTimePtr(args{hour: 15, minute: 0}),
			offTime:     testTimePtr(args{hour: 11, minute: 0}),
			isEmpty:     false,
			isOn:        true,
		},
	}

	for _, tt := range tests {
		runTableTestsTest(t, tt)
	}
}

func runTableTestsTest(t *testing.T, tt test) {
	t.Run(tt.title, func(t *testing.T) {
		if tt.skip {
			t.Skip()
		}

		is := is.NewRelaxed(t)
		onOffTimes := &OnOffTimes{On: tt.onTime, Off: tt.offTime}
		if !tt.forceEmptyWeekday && tt.onTime == nil && tt.offTime == nil {
			onOffTimes = nil
		}
		empty, onOrOff := isTimeOnOrOff(tt.currentTime, onOffTimes)
		is.Equal(tt.isEmpty, empty) // check if there's a time entry for this day
		is.Equal(tt.isOn, onOrOff)  // check if camera is on or not
	})
}
