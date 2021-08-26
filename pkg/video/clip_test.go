package video

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/log"
)

const testClipPath = "/testroot/clips/TestConn/2010-02-2/2010-02-02 19:45:00"

func overloadFatalLog(overload func(string, ...interface{})) func() {
	logFatalRef := log.Fatal
	log.Fatal = overload
	return func() { log.Fatal = logFatalRef }
}

func TestNewClip(t *testing.T) {
	clip := NewClip(testClipPath)
	assert.NotNil(t, clip)
}

type testFrame struct {
	onClose func()
}

func (frame *testFrame) DataRef() interface{} {
	return nil
}

func (frame *testFrame) Dimensions() (int, int) {
	return 100, 50
}

func (frame *testFrame) Close() {
	if frame.onClose != nil {
		frame.onClose()
	}
}

func TestClipAppendFrameTracksFrameButDoesNotCloseIt(t *testing.T) {
	clip := NewClip(testClipPath)
	require.NotNil(t, clip)

	frameCloseInvoked := false
	frame := &testFrame{onClose: func() { frameCloseInvoked = true }}
	clip.AppendFrame(frame)

	assert.Contains(t, clip.GetFrames(), frame)
	assert.False(t, frameCloseInvoked)
}

func TestClipAppendFrameTracksFrameWhichIsThenClosed(t *testing.T) {
	clip := NewClip(testClipPath)
	require.NotNil(t, clip)

	frameCloseInvoked := false
	frame := &testFrame{onClose: func() { frameCloseInvoked = true }}
	clip.AppendFrame(frame)

	assert.Contains(t, clip.GetFrames(), frame)
	clip.Close()

	assert.True(t, frameCloseInvoked)
}

func TestClipAppendFailsIfClipAlreadyClosed(t *testing.T) {
	var fatalLogs []string
	resetLogFatal := overloadFatalLog(
		func(format string, a ...interface{}) {
			fatalLogs = append(fatalLogs, fmt.Sprintf(format, a...))
		},
	)
	defer resetLogFatal()

	logging.CurrentLoggingLevel = logging.SilentLevel
	defer func() { logging.CurrentLoggingLevel = logging.WarnLevel }()

	clip := NewClip(testClipPath)
	require.NotNil(t, clip)

	clip.Close()
	clip.AppendFrame(&testFrame{})

	assert.Contains(t, fatalLogs, "cannot append frame to closed clip")
}
