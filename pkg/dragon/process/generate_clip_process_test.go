package process_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/dragon/process"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

// 30 fps * 2 seconds per clip
const framesPerClip = 60
const persistLoc = "/testroot/clips"

func TestNewGenerateClipProcess(t *testing.T) {
	frames := make(chan video.Frame)
	generatedClips := make(chan video.Clip)

	is := is.New(t)
	proc := process.NewGenerateClipProcess(frames, generatedClips, framesPerClip, persistLoc)
	is.True(proc != nil)
}
