package config

import (
	"github.com/tauraamui/dragondaemon/internal/config"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
)

type Creator interface {
	configdef.Creator
}

func DefaultCreator() Creator {
	return config.DefaultCreator()
}
