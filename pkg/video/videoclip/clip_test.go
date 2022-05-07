package videoclip_test

import (
	"fmt"
	"testing"

	"github.com/matryer/is"
	"github.com/stretchr/testify/require"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video/videoclip"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
	"github.com/tauraamui/dragondaemon/pkg/xis"
)

const testClipPath = "/testroot/clips/TestConn/2010-02-2/2010-02-02 19:45:00"

func overloadFatalLog(overload func(string, ...interface{})) func() {
	logFatalRef := log.Fatal
	log.Fatal = overload
	return func() { log.Fatal = logFatalRef }
}

func TestNewClip(t *testing.T) {
	is := is.New(t)
	clip := videoclip.New(testClipPath, 22)
	is.True(clip != nil)
}

type testFrame struct {
	onClose func()
}

func (frame *testFrame) Timestamp() int64 { return 0 }

func (frame *testFrame) DataRef() interface{} {
	return nil
}

func (frame *testFrame) Dimensions() videoframe.Dimensions {
	return videoframe.Dimensions{W: 100, H: 50}
}

func (frame *testFrame) ToBytes() []byte { return nil }

func (frame *testFrame) Close() {
	if frame.onClose != nil {
		frame.onClose()
	}
}

func TestClipAppendFrameTracksFrameButDoesNotCloseIt(t *testing.T) {
	is := is.New(t)
	clip := videoclip.New(testClipPath, 22)
	require.NotNil(t, clip)

	frameCloseInvoked := false
	frame := &testFrame{onClose: func() { frameCloseInvoked = true }}
	clip.AppendFrame(frame)

	xis := xis.New(is)
	xis.Contains(clip.Frames(), frame)
	is.True(frameCloseInvoked == false)
}

func TestClipAppendFrameTracksFrameWhichIsThenClosed(t *testing.T) {
	t.Skip("FRAMES ARE NOT BEING CLOSED AT THIS TIME")
	is := is.New(t)
	clip := videoclip.New(testClipPath, 22)
	is.True(clip != nil)

	frameCloseInvoked := false
	frame := &testFrame{onClose: func() { frameCloseInvoked = true }}
	clip.AppendFrame(frame)

	xis := xis.New(is)
	xis.Contains(clip.Frames(), frame)
	xis.Contains(clip.Frames(), frame)
	clip.Close()

	is.True(frameCloseInvoked)
}

func TestClipAppendFailsIfClipAlreadyClosed(t *testing.T) {
	is := is.New(t)
	var fatalLogs []string
	resetLogFatal := overloadFatalLog(
		func(format string, a ...interface{}) {
			fatalLogs = append(fatalLogs, fmt.Sprintf(format, a...))
		},
	)
	defer resetLogFatal()

	logging.CurrentLoggingLevel = logging.SilentLevel
	defer func() { logging.CurrentLoggingLevel = logging.WarnLevel }()

	clip := videoclip.New(testClipPath, 22)
	is.True(clip != nil)

	clip.Close()
	clip.AppendFrame(&testFrame{})

	xis := xis.New(is)
	xis.Contains(fatalLogs, "cannot append frame to closed clip")
}
