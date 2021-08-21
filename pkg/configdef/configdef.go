package configdef

import (
	"errors"
	"fmt"

	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"gopkg.in/dealancer/validate.v2"
)

type Camera struct {
	Title           string            `json:"title" validate:"empty=false"`
	Address         string            `json:"address"`
	PersistLoc      string            `json:"persist_location" validate:"empty=false"`
	MaxClipAgeDays  int               `json:"max_clip_age_days" validate:"gte=1 & lte=30"`
	MockWriter      bool              `json:"mock_writer"`
	MockCapturer    bool              `json:"mock_capturer"`
	FPS             int               `json:"fps" validate:"gte=1 & lte=30"`
	DateTimeLabel   bool              `json:"date_time_label"`
	DateTimeFormat  string            `json:"date_time_format"`
	SecondsPerClip  int               `json:"seconds_per_clip" validate:"gte=1 & lte=3"`
	Disabled        bool              `json:"disabled"`
	Schedule        schedule.Schedule `json:"schedule"`
	ReolinkAdvanced ReolinkAdvanced   `json:"reolink_advanced"`
}

type ReolinkAdvanced struct {
	Enabled    bool   `json:"enabled"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	APIAddress string `json:"api_address"`
}

type Values struct {
	Debug   bool     `json:"debug"`
	Secret  string   `json:"secret"`
	Cameras []Camera `json:"cameras"`
}

func (v Values) RunValidate() error {
	return v.runValidate()
}

func (v Values) runValidate() error {
	const validationErrorHeader = "validation failed: %w"
	if hasDupCameraTitles(v.Cameras) {
		return fmt.Errorf(validationErrorHeader, errors.New("camera titles must be unique"))
	}
	return validate.Validate(&v)
}

func hasDupCameraTitles(cameras []Camera) (hasDup bool) {
	hasDup = false
	if len(cameras) == 0 {
		return
	}

	for ci, cam := range cameras {
		for i := ci; i < len(cameras); i++ {
			if i == ci {
				continue
			}
			if cam.Title == cameras[i].Title {
				hasDup = true
				return
			}
		}
	}
	return
}
