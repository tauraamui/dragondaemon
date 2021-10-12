package schedule

import (
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/stretchr/testify/suite"
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

type ScheduleIsTimeOnOrOffTestSuite struct {
	suite.Suite
}

func TestScheduleIsTimeOnOrOffTestSuite(t *testing.T) {
	suite.Run(t, &ScheduleIsTimeOnOrOffTestSuite{})
}

type test struct {
	title       string
	currentTime Time
	onTime      *Time
	offTime     *Time

	isEmpty bool
	isOn    bool
}

func (suite *ScheduleIsTimeOnOrOffTestSuite) TestTimesMatchExpectedState() {
	t := suite.T()

	tests := []test{
		{
			title:       "current time is after nil unspecified time should be on",
			currentTime: Time(time.Now()),
			isEmpty:     true,
			isOn:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			is := is.NewRelaxed(t)
			onOffTimes := &OnOffTimes{On: tt.onTime, Off: tt.offTime}
			if tt.onTime == nil && tt.offTime == nil {
				onOffTimes = nil
			}
			empty, onOrOff := isTimeOnOrOff(tt.currentTime, onOffTimes)
			is.Equal(tt.isEmpty, empty)
			is.Equal(tt.isOn, onOrOff)
		})
	}
}

func (suite *ScheduleIsTimeOnOrOffTestSuite) TestCurrentTimeIsAfterNilUnSpecifiedTime() {
	suite.T().Skip()
	is := is.New(suite.T())
	empty, onOrOff := isTimeOnOrOff(Time(time.Now()), nil)
	is.True(empty)
	is.True(onOrOff)
}

func (suite *ScheduleIsTimeOnOrOffTestSuite) TestCurrentTimeIsBeforeOff() {
	is := is.New(suite.T())
	currentTime := testTime(args{hour: 11, minute: 0})
	offTime := testTime(args{hour: 12, minute: 0})

	empty, onOrOff := isTimeOnOrOff(currentTime, &OnOffTimes{
		Off: &offTime,
	})

	is.True(!empty)  // should be a time entry for this day
	is.True(onOrOff) // should be on
}

func (suite *ScheduleIsTimeOnOrOffTestSuite) TestCurrentTimeIsAfterOff() {
	is := is.New(suite.T())
	currentTime := testTime(args{hour: 13, minute: 0})
	offTime := testTime(args{hour: 9, minute: 0})

	empty, onOrOff := isTimeOnOrOff(currentTime, &OnOffTimes{
		Off: &offTime,
	})

	is.True(!empty)   // should be a time entry for this day
	is.True(!onOrOff) // should be off
}

func (suite *ScheduleIsTimeOnOrOffTestSuite) TestCurrentTimeIsAfterOn() {
	is := is.New(suite.T())
	currentTime := testTime(args{hour: 13, minute: 0})
	onTime := testTime(args{hour: 11, minute: 0})

	empty, onOrOff := isTimeOnOrOff(currentTime, &OnOffTimes{
		On: &onTime,
	})

	is.True(!empty)
	is.True(onOrOff) // should be on
}

func (suite *ScheduleIsTimeOnOrOffTestSuite) TestCurrentTimeIsBeforeOnAndAfterOff() {
	is := is.New(suite.T())
	currentTime := testTime(args{hour: 13, minute: 0})
	offTime := testTime(args{hour: 10, minute: 0})
	onTime := testTime(args{hour: 14, minute: 0})

	empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
		On:  &onTime,
		Off: &offTime,
	})

	is.True(!empty)   // should be a time entry for this day
	is.True(!onOrOff) // should be off
}

func (suite *ScheduleIsTimeOnOrOffTestSuite) TestCurrentTimeIsAfterOnAndBeforeOff() {
	is := is.New(suite.T())
	currentTime := testTime(args{hour: 14, minute: 0})
	offTime := testTime(args{hour: 17, minute: 0})
	onTime := testTime(args{hour: 13, minute: 0})

	empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
		On:  &onTime,
		Off: &offTime,
	})

	is.True(!empty)  // should be a time entry for this day
	is.True(onOrOff) // should be on
}

func (suite *ScheduleIsTimeOnOrOffTestSuite) TestCurrentTimeIsAfterOnAndAfterOff() {
	is := is.New(suite.T())
	currentTime := testTime(args{hour: 14, minute: 0})
	offTime := testTime(args{hour: 17, minute: 0})
	onTime := testTime(args{hour: 13, minute: 0})

	empty, onOrOff := isTimeOnOrOff(Time(currentTime), &OnOffTimes{
		On:  &onTime,
		Off: &offTime,
	})

	is.True(!empty)  // should be a time entry for this day
	is.True(onOrOff) // should be on
}
