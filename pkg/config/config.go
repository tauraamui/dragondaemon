package config

import (
	"github.com/tauraamui/dragondaemon/internal/config"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
)

type CreateResolver interface {
	configdef.CreateResolver
}

func DefaultCreateResolver() CreateResolver {
	return config.DefaultCreateResolver()
}
