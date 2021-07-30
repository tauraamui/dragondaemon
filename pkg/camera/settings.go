package camera

import (
	"github.com/tauraamui/dragondaemon/internal/config"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
)

type Settings struct {
	DateTimeFormat  string
	DateTimeLabel   bool
	FPS             int
	PersistLocation string
	Reolink         config.ReolinkAdvanced
	Schedule        schedule.Schedule
	SecondsPerClip  int
}
