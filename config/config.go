package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
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
	r       func(string) ([]byte, error)
	um      func([]byte, interface{}) error
	v       func(interface{}) error
	Debug   bool     `json:"debug"`
	Cameras []Camera `json:"cameras"`
}

func NewConfig() *Config {
	return &Config{
		r:  ioutil.ReadFile,
		um: json.Unmarshal,
		v:  validate.Validate,
	}
}

func (c *Config) Load() error {
	configPath := os.Getenv("DRAGON_DAEMON_CONFIG")
	if configPath == "" {
		configPath = "dd.config"
	}

	file, err := c.r(configPath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to read from path %s", configPath))
	}

	err = c.um(file, c)
	if err != nil {
		return errors.Wrap(err, "Parsing configuration file error")
	}

	err = c.v(c)
	if err != nil {
		return errors.Wrap(err, "Unable to validate configuration")
	}

	return nil
}
