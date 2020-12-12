package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/tacusci/logging"
	"gopkg.in/dealancer/validate.v2"
)

type Camera struct {
	Title          string   `json:"title" validate:"empty=false"`
	Address        string   `json:"address"`
	PersistLoc     string   `json:"persist_location"`
	SecondsPerClip int      `json:"seconds_per_clip"`
	Disabled       bool     `json:"disabled"`
	Schedule       Schedule `json:"schedule"`
}

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

type OnOffTimes struct {
	Off string `json:"off"`
	On  string `json:"on"`
}

type Config struct {
	Cameras []Camera
}

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
