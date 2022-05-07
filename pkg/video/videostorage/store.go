package videostorage

import "github.com/tauraamui/dragondaemon/pkg/video/videoframe"

type Storage interface {
	SaveFrames(time uint, frames []videoframe.Frame)
}
