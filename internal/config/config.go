package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"github.com/tauraamui/dragondaemon/pkg/log"
)

const (
	vendorName     = "tacusci"
	appName        = "dragondaemon"
	configFileName = "config.json"
)

var fs afero.Fs = afero.NewOsFs()

type Camera struct {
	Title           string            `json:"title" validate:"empty=false"`
	Address         string            `json:"address"`
	PersistLoc      string            `json:"persist_location"`
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
	Debug            bool     `json:"debug"`
	Secret           string   `json:"secret"`
	MaxClipAgeInDays int      `json:"max_clip_age_in_days" validate:"gte=1 & lte=30"`
	Cameras          []Camera `json:"cameras"`
}

func Load() (Values, error) {
	var values Values

	configPath, err := resolveConfigPath()
	if err != nil {
		return values, err
	}

	log.Info("Resolved config file location: %s", configPath)
	file, err := readConfigFile(configPath)
	if err != nil {
		return values, err
	}

	err = unmarshal(file, &values)
	if err != nil {
		return values, err
	}

	return values, nil
}

var readConfigFile = func(path string) ([]byte, error) {
	return afero.ReadFile(fs, path)
}

func unmarshal(content []byte, values *Values) error {
	err := json.Unmarshal(content, values)
	if err != nil {
		return errors.Errorf("parsing configuration error: %w", err)
	}
	return nil
}

func resolveConfigPath() (string, error) {
	configPath := os.Getenv("DRAGON_DAEMON_CONFIG")
	if len(configPath) > 0 {
		return configPath, nil
	}

	configParentDir, err := userConfigDir()
	if err != nil {
		return "", fmt.Errorf("unable to resolve %s config file location: %w", configFileName, err)
	}

	return filepath.Join(
		configParentDir,
		vendorName,
		appName,
		configFileName), nil
}

var userConfigDir = func() (string, error) {
	return os.UserConfigDir()
}
