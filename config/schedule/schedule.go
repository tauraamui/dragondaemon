package schedule

import (
	"fmt"
	"strings"
	"time"
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

func (st *Time) String() string {
	t := time.Time(*st)
	return fmt.Sprintf("%q", t.Format(stLayout))
}

// Schedule contains each day of the week and it's off and on time entries
type Schedule struct {
	Everyday  OnOffTimes `json:"everyday"`
	Monday    OnOffTimes `json:"monday"`
	Tuesday   OnOffTimes `json:"tuesday"`
	Wednesday OnOffTimes `json:"wednesday"`
	Thursday  OnOffTimes `json:"thursday"`
	Friday    OnOffTimes `json:"friday"`
	Saturday  OnOffTimes `json:"saturday"`
	Sunday    OnOffTimes `json:"sunday"`
}

// IsOn returns whether given time is within on period from schedule
func (s Schedule) IsOn(t Time) bool {
	switch t.Weekday().String() {
	case "Monday":
		return isTimeOnOrOff(s.Monday, t)
	case "Tuesday":
		return isTimeOnOrOff(s.Tuesday, t)
	case "Wednesday":
		return isTimeOnOrOff(s.Wednesday, t)
	case "Thursday":
		return isTimeOnOrOff(s.Thursday, t)
	case "Friday":
		return isTimeOnOrOff(s.Friday, t)
	case "Saturday":
		return isTimeOnOrOff(s.Saturday, t)
	case "Sunday":
		return isTimeOnOrOff(s.Sunday, t)
	}
	return false
}

func isTimeOnOrOff(onOff OnOffTimes, t Time) bool {
	if onOff.On != nil {
		if t.After(*onOff.On) {
			if onOff.Off == nil {
				return true
			}

			if t.After(*onOff.Off) {
				return false
			}

			return true
		}
	}

	if onOff.Off != nil {
		if t.After(*onOff.Off) {
			return false
		}

		return true
	}

	if onOff.On == nil && onOff.Off == nil {
		return true
	}

	return false
}

// OnOffTimes for loading up on off time entries
type OnOffTimes struct {
	Off *Time `json:"off"`
	On  *Time `json:"on"`
}
