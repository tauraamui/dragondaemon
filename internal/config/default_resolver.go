package config

import (
	"github.com/tauraamui/dragondaemon/pkg/configdef"
)

func DefaultResolver() configdef.Resolver {
	return defaultResolver{}
}

type defaultResolver struct{}

func (d defaultResolver) Resolve() (configdef.Values, error) {
	return load()
}
