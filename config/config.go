package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/shibukawa/configdir"
	"github.com/tacusci/logging/v2"

	"github.com/pkg/errors"
	"github.com/tauraamui/dragondaemon/config/schedule"
	"gopkg.in/dealancer/validate.v2"
)

const (
	configDirType  = configdir.System
	vendorName     = "tacusci"
	appName        = "dragondaemon"
	configFileName = "config.json"
)

var (
	configDir configdir.ConfigDir
)

func init() {
	configDir = configdir.New(vendorName, appName)
}

// Camera configuration
type Camera struct {
	Title           string            `json:"title" validate:"empty=false"`
	Address         string            `json:"address"`
	PersistLoc      string            `json:"persist_location"`
	FPS             int               `json:"fps" validate:"gte=1 & lte=30"`
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

// Config to keep track of each loaded camera's configuration
type values struct {
	r                func(string) ([]byte, error)
	um               func([]byte, interface{}) error
	v                func(interface{}) error
	Debug            bool     `json:"debug"`
	MaxClipAgeInDays int      `json:"max_clip_age_in_days" validate:"gte=1 & lte=30"`
	Cameras          []Camera `json:"cameras"`
}

func New() *values {
	return &values{
		r:  ioutil.ReadFile,
		um: json.Unmarshal,
		v:  validate.Validate,
	}
}

func (c *values) Load() error {
	configPath, err := resolveConfigPath()
	if err != nil {
		return err
	}

	logging.Info("Resolved config file location: %s", configPath)

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

func resolveConfigPath() (string, error) {
	configPath := os.Getenv("DRAGON_DAEMON_CONFIG")
	if len(configPath) == 0 {
		configParentDir := configDir.QueryFolderContainsFile(configFileName)
		if configParentDir == nil {
			return "", errors.New("unable to resolve dragondaemon.json config file location")
		}
		return fmt.Sprintf("%s%c%s", configParentDir.Path, os.PathSeparator, configFileName), nil
	}
	return configPath, nil
}
