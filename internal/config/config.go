package config

import (
	"github.com/spf13/afero"
)

var fs afero.Fs = afero.NewOsFs()
