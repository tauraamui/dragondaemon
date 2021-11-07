package video_test

import (
	"context"
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/video"
	"github.com/tauraamui/dragondaemon/pkg/video/videobackend"
	"github.com/tauraamui/dragondaemon/pkg/video/videoclip"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
)

type testVideoBackend struct {
	connectCallback func()
}

func (tvb testVideoBackend) Connect(context context.Context, address string) (videobackend.Connection, error) {
	tvb.connectCallback()
	return testVideoConnection{}, nil
}

func (tvb testVideoBackend) NewFrame() videoframe.Frame {
	return testVideoFrame{}
}

func (tvb testVideoBackend) NewWriter() videoclip.Writer {
	return nil
}

type testVideoFrame struct {
}

func (tvf testVideoFrame) DataRef() interface{} {
	return nil
}

func (tvf testVideoFrame) Dimensions() videoframe.FrameDimension {
	return videoframe.FrameDimension{W: 100, H: 50}
}

func (tvf testVideoFrame) Close() {}

type testVideoConnection struct {
}

func (tvc testVideoConnection) UUID() string {
	return "test-conn-uuid"
}

func (tvc testVideoConnection) Read(frame videoframe.Frame) error {
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
