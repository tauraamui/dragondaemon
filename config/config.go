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
	vendorName     = "tacusci"
	appName        = "dragondaemon"
	configFileName = "config.json"
)

type defaultSettingKey int

const (
	MAXCLIPAGEINDAYS defaultSettingKey = 0x0
	DATETIMEFORMAT   defaultSettingKey = 0x1
)

var (
	configDir       configdir.ConfigDir
	defaultSettings = map[defaultSettingKey]string{
		MAXCLIPAGEINDAYS: "__N_30",
		DATETIMEFORMAT:   "__S_2006/01/02 15:04:05.999999999",
	}
)

func init() {
	configDir = configdir.New(vendorName, appName)
}

func WriteDefault() error {
	configPath := os.Getenv("DRAGON_DAEMON_CONFIG")
	if len(configPath) == 0 {
		configParentDir := configDir.QueryFolders(configdir.Global)
		if len(configParentDir) > 0 {
			logging.Debug("CONFIG LOCATION: %s", configParentDir[0].Path)
		}
	}
	return nil
}

// Camera configuration
type Camera struct {
	Title           string            `json:"title" validate:"empty=false"`
	Address         string            `json:"address"`
	PersistLoc      string            `json:"persist_location"`
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

// Config to keep track of each loaded camera's configuration
type values struct {
	r                func(string) ([]byte, error)
	um               func([]byte, interface{}) error
	v                func(interface{}) error
	Debug            bool     `json:"debug"`
	Secret           string   `json:"secret"`
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

	c.loadDefaults()

	err = c.v(c)
	if err != nil {
		return errors.Wrap(err, "Unable to validate configuration")
	}

	return nil
}

func (c *values) loadDefaults() {
	for i := 0; i < len(c.Cameras); i++ {
		camera := &c.Cameras[i]
		if len(camera.DateTimeFormat) == 0 {
			camera.DateTimeFormat = defaultSettings[DATETIMEFORMAT]
		}
	}
}

func resolveConfigPath() (string, error) {
	configPath := os.Getenv("DRAGON_DAEMON_CONFIG")
	if len(configPath) == 0 {
		configParentDir := configDir.QueryFolderContainsFile(configFileName)
		if configParentDir == nil {
			return "", fmt.Errorf("unable to resolve %s config file location", configFileName)
		}
		return fmt.Sprintf("%s%c%s", configParentDir.Path, os.PathSeparator, configFileName), nil
	}
	return configPath, nil
}
