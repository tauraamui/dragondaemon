package video_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type testVideoBackend struct {
	connectCallback func()
}

func (tvb testVideoBackend) Connect(context context.Context, address string) (video.Connection, error) {
	tvb.connectCallback()
	return testVideoConnection{}, nil
}

func (tvb testVideoBackend) NewFrame() video.Frame {
	return testVideoFrame{}
}

func (tvb testVideoBackend) NewWriter() video.ClipWriter {
	return nil
}

type testVideoFrame struct {
}

func (tvf testVideoFrame) DataRef() interface{} {
	return nil
}

func (tvf testVideoFrame) Dimensions() video.FrameDimension {
	return video.FrameDimension{W: 100, H: 50}
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

func TestConnectInvokesBackendConnect(t *testing.T) {
	invokedConnect := false
	conn, err := video.Connect("fakeaddr", testVideoBackend{
		connectCallback: func() { invokedConnect = true },
	})
	require.NoError(t, err)
	assert.NotNil(t, conn)
	assert.True(t, invokedConnect)
}

func TestConnectWithCancelInvokesBackendConnect(t *testing.T) {
	invokedConnect := false
	conn, err := video.ConnectWithCancel(context.TODO(), "fakeaddr", testVideoBackend{
		connectCallback: func() { invokedConnect = true },
	})
	require.NoError(t, err)
	assert.NotNil(t, conn)
	assert.True(t, invokedConnect)
}
