package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/tacusci/logging"
	"gopkg.in/dealancer/validate.v2"
)

type Camera struct {
	Title          string `json:"title" validate:"empty=false"`
	Address        string `json:"address"`
	PersistLoc     string `json:"persist_location"`
	SecondsPerClip int    `json:"seconds_per_clip"`
	Disabled       bool   `json:"disabled"`
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
