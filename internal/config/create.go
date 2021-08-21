package config

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/tauraamui/dragondaemon/pkg/log"

	"github.com/tauraamui/dragondaemon/pkg/configdef"
)

func create() error {
	data, err := loadRawDefaultConfig()
	if err != nil {
		log.Fatal("unable to init default config into memory: %v", err)
	}

	path := mustResolveConfigPath()

	return writeConfigData(data, path, false)
}

func writeConfigData(data []byte, path string, overwrite bool) error {
	return errors.New("writing to file not implemented yet")
}

func loadRawDefaultConfig() ([]byte, error) {
	return json.MarshalIndent(
		configdef.Values{
			MaxClipAgeInDays: defaultSettings[MAXCLIPAGEINDAYS].(int),
			Cameras:          defaultSettings[CAMERAS].([]configdef.Camera),
		}, "", " ")
}

func mustResolveConfigPath() string {
	path, err := resolveConfigPath()
	if err != nil {
		log.Fatal("unable to resolve config path: %v", err)
	}

	parentDirPath := strings.Replace(path, configFileName, "", -1)
	if _, err := fs.Stat(parentDirPath); errors.Is(err, os.ErrNotExist) {
		err = fs.MkdirAll(parentDirPath, os.ModeDir|os.ModePerm)
		if err != nil {
			log.Fatal("unable to create config parent directory: %v", err)
		}
	}

	return path
}
