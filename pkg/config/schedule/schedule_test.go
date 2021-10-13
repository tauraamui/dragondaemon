package schedule

import (
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestTimeFromJSON(t *testing.T) {
	is := is.New(t)

	todayRef := TODAY
	resetToday := func() { TODAY = todayRef }
	defer resetToday()

	TODAY = time.Date(2021, 3, 1, 0, 0, 0, 0, time.UTC)
	var timeInst Time
	timeInst.UnmarshalJSON([]byte(`"14:15:19"`))

	is.Equal(timeInst.Year(), 2021)
	is.Equal(int(timeInst.Month()), 3)
	is.Equal(timeInst.Day(), 1)
	is.Equal(timeInst.Hour(), 14)
	is.Equal(timeInst.Minute(), 15)
	is.Equal(timeInst.Second(), 19)
	is.Equal(timeInst.Nanosecond(), 0)
	is.Equal(timeInst.Location(), time.UTC)
	is.Equal(timeInst.String(), `"14:15:19"`)
	is.Equal(timeInst.Weekday(), time.Monday)
}

func TestTimeSubAnotherTime(t *testing.T) {
	is := is.New(t)

	ft := testTime(args{hour: 17})
	ft.UnmarshalJSON([]byte(`"17:00:00"`))
	st := testTime(args{hour: 11})
	st.UnmarshalJSON([]byte(`"11:00:00"`))

	d := ft.Sub(time.Time(st))
	is.Equal(d.Hours(), float64(6))
}

func TestTimeMarshalJSON(t *testing.T) {
	is := is.New(t)

	timeInst := testTime(args{hour: 8, minute: 27})
	json, err := timeInst.MarshalJSON()
	is.NoErr(err)
	is.Equal(json, []byte(`"08:27:00"`))
}

type scheduleTest struct {
	skip     bool
	title    string
	today    time.Time
	schedule Week
	timeNow  time.Time
	isOn     bool
}

func TestSchedule(t *testing.T) {
	todayRef := TODAY
	resetToday := func() { TODAY = todayRef }
	defer resetToday()

	tests := []scheduleTest{
		{
			title:    "empty schedule should always be on",
			today:    time.Date(2021, 3, 17, 0, 0, 0, 0, time.UTC),
			schedule: Week{},
			isOn:     true,
		},
		{
			title: "current time after weekday with off after weekday with on should be off",
			today: time.Date(2021, 3, 17, 0, 0, 0, 0, time.UTC),
			schedule: Week{
				Monday: OnOffTimes{
					On: testTimePtr(args{hour: 21}),
				},
				Tuesday: OnOffTimes{
					Off: testTimePtr(args{hour: 10}),
				},
			},
			timeNow: time.Time(testTime(args{
				date: timeDate{2021, 3, 17},
				hour: 4, minute: 0,
			})),
			isOn: false,
		},
	}

	for _, tt := range tests {
		runIsOnOrOffWithinSchedule(t, tt)
	}
}

func runIsOnOrOffWithinSchedule(t *testing.T, tt scheduleTest) {
	t.Run(tt.title, func(t *testing.T) {
		if len(tt.title) == 0 {
			t.Error("table tests must all have titles")
		}

		if tt.skip {
			t.Skip()
		}

		is := is.NewRelaxed(t)

		TODAY = tt.today
		scheduleInst := NewSchedule(tt.schedule)
		is.Equal(scheduleInst.IsOn(Time(tt.timeNow)), tt.isOn)
	})
}

type timeDifftest struct {
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
	tests := []timeDifftest{
		{
			title: "current date+time before off after on should be on",
			onTime: testTimePtr(args{
				date: timeDate{
					2021, 3, 13,
				}, hour: 11, minute: 0,
			}),
			currentTime: testTime(args{
				date: timeDate{
					2021, 3, 15,
				}, hour: 7, minute: 0,
			}),
			offTime: testTimePtr(args{
				date: timeDate{
					2021, 3, 17,
				}, hour: 3, minute: 0,
			}),
			isEmpty: false,
			isOn:    true,
		},
		{
			title: "current date+time before on after off should be off",
			offTime: testTimePtr(args{
				date: timeDate{
					2021, 3, 17,
				}, hour: 3, minute: 0,
			}),
			currentTime: testTime(args{
				date: timeDate{
					2021, 3, 19,
				}, hour: 7, minute: 0,
			}),
			onTime: testTimePtr(args{
				date: timeDate{
					2021, 3, 21,
				}, hour: 11, minute: 0,
			}),
			isEmpty: false,
			isOn:    false,
		},
		{
			title: "current date+time after on after off should be on",
			offTime: testTimePtr(args{
				date: timeDate{
					2021, 3, 17,
				}, hour: 3, minute: 0,
			}),
			onTime: testTimePtr(args{
				date: timeDate{
					2021, 3, 18,
				}, hour: 11, minute: 0,
			}),
			currentTime: testTime(args{
				date: timeDate{
					2021, 3, 23,
				}, hour: 7, minute: 0,
			}),
			isEmpty: false,
			isOn:    true,
		},
		{
			title: "current date+time after off after on should be off",
			onTime: testTimePtr(args{
				date: timeDate{
					2021, 3, 13,
				}, hour: 11, minute: 0,
			}),
			offTime: testTimePtr(args{
				date: timeDate{
					2021, 3, 17,
				}, hour: 3, minute: 0,
			}),
			currentTime: testTime(args{
				date: timeDate{
					2021, 3, 19,
				}, hour: 7, minute: 0,
			}),
			isEmpty: false,
			isOn:    false,
		},
	}

	for _, tt := range tests {
		runIsTimeOnOrOffTest(t, tt)
	}
}

func TestSameDayScheduleTimesMatchExpectedState(t *testing.T) {
	tests := []timeDifftest{
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
		{
			title:       "current time is after off after on should be off",
			currentTime: testTime(args{hour: 20, minute: 0}),
			onTime:      testTimePtr(args{hour: 11, minute: 0}),
			offTime:     testTimePtr(args{hour: 16, minute: 0}),
			isEmpty:     false,
			isOn:        false,
		},
	}

	for _, tt := range tests {
		runIsTimeOnOrOffTest(t, tt)
	}
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

var defaultDate timeDate = timeDate{2021, 9, 13}

func timeFromHoursAndMinutes(td timeDate, hour, minute int) Time {
	return Time(time.Date(td.year, time.Month(td.month), td.day, hour, minute, 0, 0, time.UTC))
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

func runIsTimeOnOrOffTest(t *testing.T, tt timeDifftest) {
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
