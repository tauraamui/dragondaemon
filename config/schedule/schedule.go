package schedule

import (
	"fmt"
	"strings"
	"time"

	"github.com/tacusci/logging/v2"
)

type Time time.Time

var TODAY time.Time = time.Now()

const stLayout = "15:04:05"

func ParseTime(value string) (Time, error) {
	nt, err := time.Parse(stLayout, value)
	dt := time.Date(
		TODAY.Year(),
		TODAY.Month(),
		TODAY.Day(),
		nt.Hour(),
		nt.Minute(),
		nt.Second(),
		nt.Nanosecond(),
		nt.Location())
	st := Time(dt)
	return st, err
}

func (st *Time) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	nt, err := time.Parse(stLayout, s)
	dt := time.Date(
		TODAY.Year(),
		TODAY.Month(),
		TODAY.Day(),
		nt.Hour(),
		nt.Minute(),
		nt.Second(),
		nt.Nanosecond(),
		nt.Location())
	*st = Time(dt)
	return
}

func (st *Time) Sub(u time.Time) time.Duration {
	t := time.Time(*st)
	return t.Sub(u)
}

func (st *Time) MarshalJSON() ([]byte, error) {
	return []byte(st.String()), nil
}

func (st *Time) Before(u Time) bool {
	t := time.Time(*st)
	return t.Before(time.Time(u))
}

func (st *Time) After(u Time) bool {
	t := time.Time(*st)
	return t.After(time.Time(u))
}

func (st *Time) Weekday() time.Weekday {
	t := time.Time(*st)
	return t.Weekday()
}

func (st *Time) Hour() int {
	t := time.Time(*st)
	return t.Hour()
}

func (st *Time) Minute() int {
	t := time.Time(*st)
	return t.Hour()
}

func (st *Time) Second() int {
	t := time.Time(*st)
	return t.Second()
}

func (st *Time) Nanosecond() int {
	t := time.Time(*st)
	return t.Nanosecond()
}

func (st *Time) Location() *time.Location {
	t := time.Time(*st)
	return t.Location()
}

func (st *Time) String() string {
	t := time.Time(*st)
	return fmt.Sprintf("%q", t.Format(stLayout))
}

// Schedule contains each day of the week and it's off and on time entries
type Schedule struct {
	hasRunSetup            bool
	weekdayStringToWeekDay map[string]*OnOffTimes
	Everyday               OnOffTimes `json:"everyday"`
	Monday                 OnOffTimes `json:"monday"`
	Tuesday                OnOffTimes `json:"tuesday"`
	Wednesday              OnOffTimes `json:"wednesday"`
	Thursday               OnOffTimes `json:"thursday"`
	Friday                 OnOffTimes `json:"friday"`
	Saturday               OnOffTimes `json:"saturday"`
	Sunday                 OnOffTimes `json:"sunday"`
}

func (s *Schedule) setupState() {
	if s.hasRunSetup {
		return
	}
	s.weekdayStringToWeekDay = map[string]*OnOffTimes{
		"Monday":    &s.Monday,
		"Tuesday":   &s.Tuesday,
		"Wednesday": &s.Wednesday,
		"Thursday":  &s.Thursday,
		"Friday":    &s.Friday,
		"Saturday":  &s.Saturday,
		"Sunday":    &s.Sunday,
	}

	// from today to a week before set each weekday time to have relative date
	for i := 0; i < 7; i++ {
		previousDay := TODAY.AddDate(0, 0, i*-1)
		logging.Debug("Setting relative date from TODAY for %s", previousDay.Weekday().String())
		previousDayRef := s.weekdayStringToWeekDay[previousDay.Weekday().String()]
		if previousDayRef.On != nil {
			*previousDayRef.On = Time(
				time.Date(
					previousDay.Year(),
					previousDay.Month(),
					previousDay.Day(),
					previousDayRef.On.Hour(),
					previousDayRef.On.Minute(),
					previousDayRef.On.Second(),
					previousDayRef.On.Nanosecond(),
					previousDayRef.On.Location(),
				),
			)
		}

		if previousDayRef.Off != nil {
			*previousDayRef.Off = Time(
				time.Date(
					previousDay.Year(),
					previousDay.Month(),
					previousDay.Day(),
					previousDayRef.Off.Hour(),
					previousDayRef.Off.Minute(),
					previousDayRef.Off.Second(),
					previousDayRef.Off.Nanosecond(),
					previousDayRef.Off.Location(),
				),
			)
		}
	}
	s.hasRunSetup = true
}

// IsOn returns whether given time is within on period from schedule
func (s *Schedule) IsOn(t Time) bool {
	s.setupState()

	for i := 0; i < 7; i++ {
		previousDay := TODAY.AddDate(0, 0, i*-1)
		previousDayRef := s.weekdayStringToWeekDay[previousDay.Weekday().String()]
		empty, state := timeAfterOnOrOff(t, previousDayRef)
		if !empty {
			return state
		}
	}

	return true
}

func timeAfterOnOrOff(t Time, weekday *OnOffTimes) (empty bool, state bool) {
	if weekday.On != nil && weekday.Off != nil {
		if weekday.On.After(*weekday.Off) {
			if t.After(*weekday.On) {
				return false, true
			}
		}

		if weekday.Off.After(*weekday.On) {
			if t.After(*weekday.Off) {
				return false, false
			}
		}
	}

	if weekday.On == nil && weekday.Off != nil {
		if t.After(*weekday.Off) {
			return false, false
		}
	}

	if weekday.On != nil && weekday.Off == nil {
		if t.After(*weekday.On) {
			return false, true
		}
	}

	return true, true
}

func isTimeOnOrOff(onOff OnOffTimes, t Time) bool {
	if onOff.On != nil && onOff.Off == nil {
		if t.After(*onOff.On) {
			return true
		}
	}

	if onOff.Off != nil && onOff.On == nil {
		if t.After(*onOff.Off) {
			return false
		}
	}

	if onOff.On != nil && onOff.Off != nil {
		if onOff.On.After(*onOff.Off) {
			if t.After(*onOff.On) {
				return true
			}
		}

		if onOff.Off.After(*onOff.On) {
			if t.After(*onOff.Off) {
				return false
			}
		}
	}

	return true
}

// OnOffTimes for loading up on off time entries
type OnOffTimes struct {
	Off       *Time `json:"off"`
	On        *Time `json:"on"`
	weekIndex int
}
