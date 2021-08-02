package config

import (
	"github.com/tauraamui/dragondaemon/internal/config"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
)

type Resolver interface {
	Load() (configdef.Values, error)
}

func DefaultResolver() Resolver {
	return defaultResolver{}
}

type defaultResolver struct{}

func (d defaultResolver) Load() (configdef.Values, error) {
	return config.Load()
}
