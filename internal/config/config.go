package config

import (
	"github.com/spf13/afero"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
)

var fs afero.Fs = afero.NewOsFs()

type defaultCreateResolver struct {
	creator  defaultCreator
	resolver defaultResolver
}

func DefaultCreateResolver() configdef.CreateResolver {
	return defaultCreateResolver{
		creator:  defaultCreator{},
		resolver: defaultResolver{},
	}
}

func (d defaultCreateResolver) Create() error {
	return d.creator.Create()
}

func (d defaultCreateResolver) Resolve() (configdef.Values, error) {
	return d.resolver.Resolve()
}
