package config

import "github.com/tauraamui/dragondaemon/pkg/configdef"

type defaultSettingKey uint

const (
	MAXCLIPAGEINDAYS defaultSettingKey = 0x0
	CAMERAS          defaultSettingKey = 0x1
	DATETIMEFORMAT   defaultSettingKey = 0x2
)

var defaultSettings = map[defaultSettingKey]interface{}{
	MAXCLIPAGEINDAYS: 30,
	CAMERAS:          []configdef.Camera{},
	DATETIMEFORMAT:   "2006/01/02 15:04:05.999999999",
}
