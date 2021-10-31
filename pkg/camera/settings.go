package camera

import (
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
)

type Settings struct {
	DateTimeFormat  string
	DateTimeLabel   bool
	FPS             int
	PersistLocation string
	MaxClipAgeDays  int
	Reolink         configdef.ReolinkAdvanced
	Schedule        schedule.Schedule
	SecondsPerClip  int
}
