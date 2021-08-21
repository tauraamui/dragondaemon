package config

import (
	"github.com/spf13/afero"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
)

var fs afero.Fs = afero.NewOsFs()

func Setup() error {
	return setup()
}

func Load() (configdef.Values, error) {
	return load()
}
