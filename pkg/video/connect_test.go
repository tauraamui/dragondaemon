package video_test

import (
	"context"
	"testing"

	"github.com/matryer/is"
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

func (tvc testVideoConnection) UUID() string {
	return "test-conn-uuid"
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
	is := is.New(t)
	invokedConnect := false
	conn, err := video.Connect("fakeaddr", testVideoBackend{
		connectCallback: func() { invokedConnect = true },
	})
	is.NoErr(err)
	is.True(conn != nil)
	is.True(invokedConnect)
}

func TestConnectWithCancelInvokesBackendConnect(t *testing.T) {
	is := is.New(t)
	invokedConnect := false
	conn, err := video.ConnectWithCancel(context.TODO(), "fakeaddr", testVideoBackend{
		connectCallback: func() { invokedConnect = true },
	})
	is.NoErr(err)
	is.True(conn != nil)
	is.True(invokedConnect)
}
