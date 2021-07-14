package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/afero"
	"github.com/tacusci/logging/v2"

	"github.com/pkg/errors"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
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
	ErrCreateConfigFile    = errors.New("unable to create default config file")
	ErrConfigAlreadyExists = errors.New("config file already exists")
	defaultSettings        = map[defaultSettingKey]interface{}{
		MAXCLIPAGEINDAYS: 30,
		CAMERAS:          []Camera{},
		DATETIMEFORMAT:   "2006/01/02 15:04:05.999999999",
	}
)

// Camera configuration
type Camera struct {
	Title           string            `json:"title" validate:"empty=false"`
	Address         string            `json:"address"`
	PersistLoc      string            `json:"persist_location"`
	MockWriter      bool              `json:"mock_writer"`
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
	fs               afero.Fs
	uc               func() (string, error)
	w                func(string, []byte, fs.FileMode) error
	Debug            bool     `json:"debug"`
	Secret           string   `json:"secret"`
	MaxClipAgeInDays int      `json:"max_clip_age_in_days" validate:"gte=1 & lte=30"`
	Cameras          []Camera `json:"cameras"`
}

func New() *values {
	return &values{
		fs: afero.NewOsFs(),
		uc: os.UserConfigDir,
		w:  ioutil.WriteFile,
	}
}

func Setup() error {
	c := New()
	c.ResetToDefaults()
	configPath, err := c.Save(false)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
		return fmt.Errorf("%w: %s", ErrConfigAlreadyExists, configPath)
	}

	logging.Info("Created default config at: %s", configPath) //nolint

	return nil
}

func (c *values) Save(overwrite bool) (string, error) {
	path, err := resolveConfigPath(c.uc)
	if err != nil {
		return "", err
	}

	parentDirPath := strings.Replace(path, configFileName, "", -1)
	if _, err := c.fs.Stat(parentDirPath); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(parentDirPath, os.ModeDir|os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("unable to create parent directory: %w", err)
		}
	}

	marshalledConfig, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		// this should be impossible to happen, so is not covered in tests
		return "", err
	}

	openingFlags := os.O_RDWR | os.O_CREATE
	// if we're not overwriting make open file return error if file exists
	if !overwrite {
		openingFlags |= os.O_EXCL
	}

	f, err := c.fs.OpenFile(path, openingFlags, 0666)
	if err != nil {
		return path, fmt.Errorf("unable to open file: %w", err)
	}
	defer f.Close()

	writtenBytesCount, err := f.Write(marshalledConfig)
	if err != nil {
		return "", err
	}

	if len(marshalledConfig) != writtenBytesCount {
		return "", errors.New("unable to write full config JSON to file")
	}

	return path, nil
}

func (c *values) Load() error {
	configPath, err := resolveConfigPath(c.uc)
	if err != nil {
		return err
	}

	logging.Info("Resolved config file location: %s", configPath) //nolint

	file, err := afero.ReadFile(c.fs, configPath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to read from path %s", configPath))
	}

	err = json.Unmarshal(file, c)
	if err != nil {
		return errors.Wrap(err, "Parsing configuration file error")
	}

	c.loadDefaultCameraDateLabelFormats()

	err = validate.Validate(c)
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

func resolveConfigPath(uc func() (string, error)) (string, error) {
	configPath := os.Getenv("DRAGON_DAEMON_CONFIG")
	if len(configPath) > 0 {
		return configPath, nil
	}

	configParentDir, err := uc()
	if err != nil {
		return "", fmt.Errorf("unable to resolve %s config file location: %w", configFileName, err)
	}

	return filepath.Join(
		configParentDir,
		vendorName,
		appName,
		configFileName), nil
}
