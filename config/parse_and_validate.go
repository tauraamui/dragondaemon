package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/tacusci/logging"
	"gopkg.in/dealancer/validate.v2"
)

// Camera configuration
type Camera struct {
	Title          string   `json:"title" validate:"empty=false"`
	Address        string   `json:"address"`
	PersistLoc     string   `json:"persist_location"`
	SecondsPerClip int      `json:"seconds_per_clip" validate:"gte=1 & lte=3"`
	Disabled       bool     `json:"disabled"`
	Schedule       Schedule `json:"schedule"`
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

// OnOffTimes for loading up on off time entries
type OnOffTimes struct {
	Off string `json:"off"`
	On  string `json:"on"`
}

// Config to keep track of each loaded camera's configuration
type Config struct {
	Cameras []Camera
}

// Load parses configuration file and loads settings
func Load() Config {
	file, err := ioutil.ReadFile("dd.config")
	if err != nil {
		logging.ErrorAndExit(err.Error())
	}

	cfg := Config{}
	err = json.Unmarshal(file, &cfg.Cameras)
	if err != nil {
		logging.ErrorAndExit(
			fmt.Sprintf("Error parsing dd.config: %v", err),
		)
	}

	err = validate.Validate(&cfg)
	if err != nil {
		logging.ErrorAndExit(
			fmt.Sprintf("Error validating dd.config content: %v", err),
		)
	}

	return cfg
}
