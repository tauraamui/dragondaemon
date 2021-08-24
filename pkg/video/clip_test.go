package video

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testClipPath = "/testroot/clips/TestConn/2010-02-2/2010-02-02 19:45:00"

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

	assert.Contains(t, clip.frames, frame)
	assert.False(t, frameCloseInvoked)
}

func TestClipAppendFrameTracksFrameWhichIsThenClosed(t *testing.T) {
	clip := NewClip(testClipPath)
	require.NotNil(t, clip)

	frameCloseInvoked := false
	frame := &testFrame{onClose: func() { frameCloseInvoked = true }}
	clip.AppendFrame(frame)

	assert.Contains(t, clip.frames, frame)
	clip.Close()

	assert.True(t, frameCloseInvoked)
}
