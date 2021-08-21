package config

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/log"
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
