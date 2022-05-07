package camera_test

import (
	"context"
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/video/videobackend"
	"github.com/tauraamui/dragondaemon/pkg/video/videoclip"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
	"github.com/tauraamui/xerror"
)

type testVideoBackend struct {
	onConnectError        error
	onConnectionReadError error
}

func (tvb testVideoBackend) Connect(context context.Context, address string) (videobackend.Connection, error) {
	if tvb.onConnectError != nil {
		return nil, tvb.onConnectError
	}
	return testVideoConnection{
		onReadError: tvb.onConnectionReadError,
	}, nil
}

func (tvb testVideoBackend) NewFrame() videoframe.Frame {
	return testVideoFrame{}
}

func (tvb testVideoBackend) NewWriter() videoclip.Writer {
	return nil
}

type testVideoFrame struct {
}

func (tvf testVideoFrame) Timestamp() int64 { return 0 }

func (tvf testVideoFrame) DataRef() interface{} {
	return nil
}

func (tvf testVideoFrame) Dimensions() videoframe.Dimensions {
	return videoframe.Dimensions{W: 100, H: 50}
}

func (tvf testVideoFrame) ToBytes() []byte { return nil }

func (tvf testVideoFrame) Close() {}

type testVideoConnection struct {
	onReadError error
}

func (tvc testVideoConnection) UUID() string {
	return "test-conn-uuid"
}

func (tvc testVideoConnection) Read(frame videoframe.Frame) error {
	return tvc.onReadError
}

func (tvc testVideoConnection) IsOpen() bool {
	return true
}

func (tvc testVideoConnection) Close() error {
	return nil
}

func TestConnectReturnsConnectionAndNoError(t *testing.T) {
	is := is.New(t)
	conn, err := camera.Connect("FakeCamera", "fakeaddr", camera.Settings{
		FPS:            22,
		SecondsPerClip: 3,
	}, testVideoBackend{})

	is.NoErr(err)
	is.True(conn != nil)

	is.True(len(conn.UUID()) > 0)
	is.Equal(conn.Title(), "FakeCamera")
	is.Equal(conn.FPS(), 22)
	is.Equal(conn.SPC(), 3)
	is.True(conn.IsOpen())
	is.True(conn.IsClosing() == false)
	is.NoErr(conn.Close())
	is.True(conn.IsClosing())
}

func TestConnectWithCancelReturnsConnectionAndNoError(t *testing.T) {
	is := is.New(t)
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	conn, err := camera.ConnectWithCancel(ctx, "FakeCamera", "fakeaddr", camera.Settings{
		FPS:            22,
		SecondsPerClip: 3,
	}, testVideoBackend{})

	is.NoErr(err)
	is.True(conn != nil)

	is.True(len(conn.UUID()) > 0)
	is.Equal(conn.Title(), "FakeCamera")
	is.Equal(conn.FPS(), 22)
	is.Equal(conn.SPC(), 3)
	is.True(conn.IsOpen())
	is.True(conn.IsClosing() == false)
	is.NoErr(conn.Close())
	is.True(conn.IsClosing())
}

func TestConnectReturnsNoConnectionAndError(t *testing.T) {
	is := is.New(t)
	conn, err := camera.Connect("FakeCamera", "fakeaddr", camera.Settings{}, testVideoBackend{
		onConnectError: xerror.New("test error"),
	})
	is.Equal(err.Error(), "Unable to connect to camera [FakeCamera]: test error")
	is.True(conn == nil)
}

func TestConnectReadReturnsFrameAndNoError(t *testing.T) {
	is := is.New(t)
	conn, err := camera.Connect("FakeCamera", "fakeaddr", camera.Settings{}, testVideoBackend{})

	is.NoErr(err)
	is.True(conn != nil)

	frame, err := conn.Read()
	is.NoErr(err)
	is.True(frame != nil)
}

func TestConnectReadReturnsNoFrameAndError(t *testing.T) {
	is := is.New(t)
	conn, err := camera.Connect("FakeCamera", "fakeaddr", camera.Settings{}, testVideoBackend{
		onConnectionReadError: xerror.New("test error"),
	})

	is.NoErr(err)
	is.True(conn != nil)

	frame, err := conn.Read()
	is.Equal(err.Error(), "unable to read frame from connection: test error")
	is.True(frame == nil)
}
