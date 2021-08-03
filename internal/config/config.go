package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/log"
)

const (
	vendorName     = "tacusci"
	appName        = "dragondaemon"
	configFileName = "config.json"
)

var fs afero.Fs = afero.NewOsFs()

func Load() (configdef.Values, error) {
	var values configdef.Values

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

func unmarshal(content []byte, values *configdef.Values) error {
	err := json.Unmarshal(content, values)
	if err != nil {
		return errors.Errorf("parsing configuration error: %v", err)
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
