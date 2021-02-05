package schedule

import (
	"fmt"
	"strings"
	"time"
)

type Time time.Time

const stLayout = "15:04:05"

func (st *Time) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	nt, err := time.Parse(stLayout, s)
	*st = Time(nt)
	return
}

func (st *Time) MarshalJSON() ([]byte, error) {
	return []byte(st.String()), nil
}

func ParseTime(value string) (Time, error) {
	nt, err := time.Parse(stLayout, value)
	st := Time(nt)
	return st, err
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

func (s Schedule) IsOn(t time.Time) bool {
	switch t.Weekday().String() {
	case "Monday":
		if s.Monday.On == nil && s.Monday.Off == nil {
			return false
		}
	case "Tuesday":
		if s.Tuesday.On == nil && s.Tuesday.Off == nil {
			return false
		}
	case "Wednesday":
		if s.Wednesday.On == nil && s.Wednesday.Off == nil {
			return false
		}
	case "Thursday":
		if s.Thursday.On == nil && s.Thursday.Off == nil {
			return false
		}
	case "Friday":
		if s.Friday.On == nil && s.Friday.Off == nil {
			return false
		}
	case "Saturday":
		if s.Saturday.On == nil && s.Saturday.Off == nil {
			return false
		}
	case "Sunday":
		if s.Sunday.On == nil && s.Sunday.Off == nil {
			return false
		}
	}
	return false
}

// OnOffTimes for loading up on off time entries
type OnOffTimes struct {
	Off *Time `json:"off"`
	On  *Time `json:"on"`
}
