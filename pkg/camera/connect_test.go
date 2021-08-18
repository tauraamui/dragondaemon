package camera_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type testVideoBackend struct {
	connectCallback func()
}

func (tvb testVideoBackend) Connect(context context.Context, address string) (video.Connection, error) {
	if tvb.connectCallback != nil {
		tvb.connectCallback()
	}
	return testVideoConnection{}, nil
}

func (tvb testVideoBackend) NewFrame() video.Frame {
	return testVideoFrame{}
}

type testVideoFrame struct {
}

func (tvf testVideoFrame) DataRef() interface{} {
	return nil
}

func (tvf testVideoFrame) Close() {}

type testVideoConnection struct {
}

func (tvc testVideoConnection) Read(frame video.Frame) error {
	return nil
}

func (tvc testVideoConnection) IsOpen() bool {
	return true
}

func (tvc testVideoConnection) Close() error {
	return nil
}

func TestConnectInvokesVideoConnect(t *testing.T) {
	conn, err := camera.Connect("FakeCamera", "fakeaddr", camera.Settings{
		FPS:            22,
		SecondsPerClip: 3,
	}, testVideoBackend{})
	require.NoError(t, err)
	require.NotNil(t, conn)

	assert.NotEmpty(t, conn.UUID())
	assert.Equal(t, conn.Title(), "FakeCamera")
	assert.Equal(t, conn.FPS(), 22)
	assert.Equal(t, conn.SPC(), 3)
	assert.True(t, conn.IsOpen())
	assert.False(t, conn.IsClosing())
	require.NoError(t, conn.Close())
	assert.True(t, conn.IsClosing())
}
