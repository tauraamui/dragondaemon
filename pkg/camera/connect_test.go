package camera_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type testVideoBackend struct {
	onConnectError        error
	onConnectionReadError error
}

func (tvb testVideoBackend) Connect(context context.Context, address string) (video.Connection, error) {
	if tvb.onConnectError != nil {
		return nil, tvb.onConnectError
	}
	return testVideoConnection{
		onReadError: tvb.onConnectionReadError,
	}, nil
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
	onReadError error
}

func (tvc testVideoConnection) Read(frame video.Frame) error {
	return tvc.onReadError
}

func (tvc testVideoConnection) IsOpen() bool {
	return true
}

func (tvc testVideoConnection) Close() error {
	return nil
}

func TestConnectReturnsConnectionAndNoError(t *testing.T) {
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

func TestConnectReturnsNoConnectionAndError(t *testing.T) {
	conn, err := camera.Connect("FakeCamera", "fakeaddr", camera.Settings{}, testVideoBackend{
		onConnectError: errors.New("test error"),
	})
	assert.EqualError(t, err, "Unable to connect to camera [FakeCamera]: test error")
	assert.Nil(t, conn)
}

func TestConnectReadReturnsFrameAndNoError(t *testing.T) {
	conn, err := camera.Connect("FakeCamera", "fakeaddr", camera.Settings{}, testVideoBackend{})
	require.NoError(t, err)
	require.NotNil(t, conn)

	frame, err := conn.Read()
	assert.NoError(t, err)
	assert.NotNil(t, frame)
}

func TestConnectReadReturnsNoFrameAndError(t *testing.T) {
	conn, err := camera.Connect("FakeCamera", "fakeaddr", camera.Settings{}, testVideoBackend{
		onConnectionReadError: errors.New("test error"),
	})
	require.NoError(t, err)
	require.NotNil(t, conn)

	frame, err := conn.Read()
	assert.EqualError(t, err, "unable to read frame from connection: test error")
	assert.Nil(t, frame)
}
