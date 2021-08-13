package process

import (
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

func Stream(cam camera.Connection, frames chan video.Frame) {
	stream(cam, frames)
}
