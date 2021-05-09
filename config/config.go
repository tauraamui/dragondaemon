package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"sync"

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
	CAMERAS          defaultSettingKey = 0x1
	DATETIMEFORMAT   defaultSettingKey = 0x2
)

var (
	configDir       configdir.ConfigDir
	defaultSettings = map[defaultSettingKey]interface{}{
		MAXCLIPAGEINDAYS: 30,
		CAMERAS:          []Camera{},
		DATETIMEFORMAT:   "2006/01/02 15:04:05.999999999",
	}
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
	of               func(string, int, fs.FileMode) (*os.File, error)
	w                func(string, []byte, fs.FileMode) error
	r                func(string) ([]byte, error)
	m                func(interface{}) ([]byte, error)
	um               func([]byte, interface{}) error
	v                func(interface{}) error
	Debug            bool     `json:"debug"`
	Secret           string   `json:"secret"`
	MaxClipAgeInDays int      `json:"max_clip_age_in_days" validate:"gte=1 & lte=30"`
	Cameras          []Camera `json:"cameras"`
}

func New() *values {
	return &values{
		of: os.OpenFile,
		w:  ioutil.WriteFile,
		r:  ioutil.ReadFile,
		m:  json.Marshal,
		um: json.Unmarshal,
		v:  validate.Validate,
	}
}

func (c *values) Save(overwrite bool) error {
	marshalledConfig, err := c.m(c)
	if err != nil {
		return err
	}

	configPath, err := resolveConfigPath()
	if err != nil {
		return err
	}

	openingFlags := os.O_RDWR | os.O_CREATE
	// if we're not overwriting make open file return error if file exists
	if !overwrite {
		openingFlags |= os.O_EXCL
	}

	f, err := c.of(configPath, openingFlags, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	writtenBytesCount, err := f.Write(marshalledConfig)
	if err != nil {
		return err
	}

	if len(marshalledConfig) != writtenBytesCount {
		return errors.New("unable to write full config JSON to file")
	}

	return nil
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

	c.loadDefaultCameraDateLabelFormats()

	err = c.v(c)
	if err != nil {
		return errors.Wrap(err, "Unable to validate configuration")
	}

	return nil
}

func (c *values) ResetToDefaults() {
	c.loadDefaults()
}

func (c *values) loadDefaultCameraDateLabelFormats() {
	wg := sync.WaitGroup{}
	for i := 0; i < len(c.Cameras); i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, camera *Camera) {
			defer wg.Done()
			if len(camera.DateTimeFormat) == 0 {
				camera.DateTimeFormat = defaultSettings[DATETIMEFORMAT].(string)
			}
		}(&wg, &c.Cameras[i])
	}
	wg.Wait()
}

func (c *values) loadDefaults() {
	c.MaxClipAgeInDays = defaultSettings[MAXCLIPAGEINDAYS].(int)
	c.Cameras = defaultSettings[CAMERAS].([]Camera)
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
