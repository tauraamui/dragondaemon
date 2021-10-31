package config

import (
	"github.com/tauraamui/dragondaemon/internal/config"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
)

type Resolver interface {
	configdef.Resolver
}

func DefaultResolver() Resolver {
	return config.DefaultResolver()
}
