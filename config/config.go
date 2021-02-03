package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/tacusci/logging/v2"
	"gopkg.in/dealancer/validate.v2"
)

// Camera configuration
type Camera struct {
	Title          string   `json:"title" validate:"empty=false"`
	Address        string   `json:"address"`
	PersistLoc     string   `json:"persist_location"`
	FPS            int      `json:"fps" validate:"gte=1 & lte=30"`
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
	Debug   bool     `json:"debug"`
	Cameras []Camera `json:"cameras"`
}

// Load parses configuration file and loads settings
func Load() Config {
	configPath := os.Getenv("DRAGON_DAEMON_CONFIG")
	if configPath == "" {
		configPath = "dd.config"
	}

	logging.Info("Loading configuration: %s", configPath)
	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		logging.Fatal(err.Error())
	}

	logging.Info("Loaded configuration...")

	cfg := Config{}
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		logging.Fatal("Error passing dd.config: %v", err)
	}

	err = validate.Validate(&cfg)
	if err != nil {
		logging.Fatal("Error validation dd.config content: %v", err)
	}

	return cfg
}
