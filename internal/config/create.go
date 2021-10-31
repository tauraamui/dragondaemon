package config

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/xerror"

	"github.com/tauraamui/dragondaemon/pkg/configdef"
)

func create() error {
	data, err := loadRawDefaultConfig()
	if err != nil {
		log.Fatal("unable to init default config into memory: %v", err)
	}

	path := mustResolveConfigPath()

	err = writeConfigToDisk(data, path, false)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return configdef.ErrConfigAlreadyExists
		}
		return err
	}

	return nil
}

func writeConfigToDisk(data []byte, path string, overwrite bool) error {
	flags := os.O_RDWR | os.O_CREATE
	if !overwrite {
		flags |= os.O_EXCL
	}

	file, err := fs.OpenFile(path, flags, 0666)
	if err != nil {
		return xerror.Errorf("unable to create/open file: %w", err)
	}
	defer file.Close()

	bc, err := file.Write(data)
	if err != nil {
		return xerror.Errorf("unable to write config to file: %s: %w", path, err)
	}

	if bc != len(data) {
		return xerror.Errorf("unable to write full config data to file: %s: %w", path, err)
	}

	return nil
}

func loadRawDefaultConfig() ([]byte, error) {
	return json.MarshalIndent(
		configdef.Values{
			Cameras: defaultSettings[CAMERAS].([]configdef.Camera),
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
