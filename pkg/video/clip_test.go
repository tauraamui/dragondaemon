package video

import (
	"fmt"
	"testing"

	"github.com/matryer/is"
	"github.com/stretchr/testify/require"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/log"
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
	clip := NewClip(testClipPath, 22)
	is.True(clip != nil)
}

type testFrame struct {
	onClose func()
}

func (frame *testFrame) DataRef() interface{} {
	return nil
}

func (frame *testFrame) Dimensions() FrameDimension {
	return FrameDimension{W: 100, H: 50}
}

func (frame *testFrame) Close() {
	if frame.onClose != nil {
		frame.onClose()
	}
}

func TestClipAppendFrameTracksFrameButDoesNotCloseIt(t *testing.T) {
	is := is.New(t)
	clip := NewClip(testClipPath, 22)
	require.NotNil(t, clip)

	frameCloseInvoked := false
	frame := &testFrame{onClose: func() { frameCloseInvoked = true }}
	clip.AppendFrame(frame)

	is.True(xis.Contains(clip.GetFrames(), frame))
	is.True(frameCloseInvoked == false)
}

func TestClipAppendFrameTracksFrameWhichIsThenClosed(t *testing.T) {
	is := is.New(t)
	clip := NewClip(testClipPath, 22)
	require.NotNil(t, clip)

	frameCloseInvoked := false
	frame := &testFrame{onClose: func() { frameCloseInvoked = true }}
	clip.AppendFrame(frame)

	is.True(xis.Contains(clip.GetFrames(), frame))
	is.True(xis.Contains(clip.GetFrames(), frame))
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

	clip := NewClip(testClipPath, 22)
	require.NotNil(t, clip)

	clip.Close()
	clip.AppendFrame(&testFrame{})

	is.True(xis.Contains(fatalLogs, "cannot append frame to closed clip"))
}
