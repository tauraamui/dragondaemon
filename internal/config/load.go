package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/xerror"
)

const (
	vendorName     = "tacusci"
	appName        = "dragondaemon"
	configFileName = "config.json"
)

func load() (configdef.Values, error) {
	var values configdef.Values

	configPath, err := resolveConfigPath()
	if err != nil {
		return configdef.Values{}, err
	}

	log.Info("Resolved config file location: %s", configPath)
	file, err := readConfigFile(configPath)
	if err != nil {
		return configdef.Values{}, err
	}

	if err := unmarshal(file, &values); err != nil {
		return configdef.Values{}, err
	}

	if err = values.RunValidate(); err != nil {
		return configdef.Values{}, err
	}

	loadDefaultCameraDataLabelFormats(values.Cameras)

	return values, nil
}

func loadDefaultCameraDataLabelFormats(cameras []configdef.Camera) {
	wg := sync.WaitGroup{}
	for i := 0; i < len(cameras); i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, camera *configdef.Camera) {
			defer wg.Done()
			if len(camera.DateTimeFormat) == 0 {
				camera.DateTimeFormat = defaultSettings[DATETIMEFORMAT].(string)
			}
		}(&wg, &cameras[i])
	}
	wg.Wait()
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
		return "", xerror.Errorf("unable to resolve %s location: %w", configFileName, err)
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
