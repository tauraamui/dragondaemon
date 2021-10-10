package config

import (
	"github.com/tauraamui/dragondaemon/internal/config"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
)

type Destoryer interface {
	configdef.Destroyer
}

func DefaultDestroyer() Destoryer {
	return config.DefaultDestoryer()
}
