package video

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gocv.io/x/gocv"
)

func overloadOpenVidCap(overload func(addr string) (*gocv.VideoCapture, error)) func() {
	openVidCapRef := openVideoCapture
	openVideoCapture = overload
	return func() { openVideoCapture = openVidCapRef }
}

func overloadReadFromVidCap(overload func(vc *gocv.VideoCapture, mat *gocv.Mat) bool) func() {
	readFromVidCapRef := readFromVideoConnection
	readFromVideoConnection = overload
	return func() { readFromVideoConnection = readFromVidCapRef }
}

func restoreMp4File() (string, error) {
	mp4Dir := os.TempDir()
	err := RestoreAsset(mp4Dir, "small.mp4")
	if err != nil {
		return "", err
	}

	return filepath.Join(mp4Dir, "small.mp4"), nil
}

func TestBackendConnect(t *testing.T) {
	mp4FilePath, err := restoreMp4File()
	require.NoError(t, err)
	defer func() { os.Remove(mp4FilePath) }()

	backend := openCVBackend{}
	conn, err := backend.Connect(context.TODO(), mp4FilePath)
	require.NoError(t, err)
	require.NotNil(t, conn)
	err = conn.Close()
	assert.Nil(t, err)
}

func TestBackendConnectWithImmediateCancelInvoke(t *testing.T) {
	mp4FilePath, err := restoreMp4File()
	require.NoError(t, err)
	defer func() { os.Remove(mp4FilePath) }()

	backend := openCVBackend{}

	ctx, cancel := context.WithCancel(context.TODO())
	errChan := make(chan error)
	go func(ctx context.Context) {
		_, err := backend.Connect(ctx, mp4FilePath)
		errChan <- err
	}(ctx)
	cancel()

	connErr := <-errChan
	assert.EqualError(t, connErr, "connection cancelled")
}

func TestConnectWithImmediateCancelInvoke(t *testing.T) {
	mp4FilePath, err := restoreMp4File()
	require.NoError(t, err)
	defer func() { os.Remove(mp4FilePath) }()

	conn := openCVConnection{}

	ctx, cancel := context.WithCancel(context.TODO())
	errChan := make(chan error)
	go func(ctx context.Context) {
		errChan <- conn.connect(ctx, mp4FilePath)
	}(ctx)
	cancel()

	connErr := <-errChan
	assert.EqualError(t, connErr, "connection cancelled")
}

func TestOpenVideoStreamInvokesOpenVideoCapture(t *testing.T) {
	resetOpenVidCap := overloadOpenVidCap(
		func(addr string) (*gocv.VideoCapture, error) {
			return nil, errors.New("test connect error")
		},
	)
	defer resetOpenVidCap()

	conn := openCVConnection{}
	err := conn.connect(context.TODO(), "TestAddr")
	assert.EqualError(t, err, "test connect error")
}

func TestOpenAndCloseVideoStream(t *testing.T) {
	mp4FilePath, err := restoreMp4File()
	require.NoError(t, err)
	defer func() { os.Remove(mp4FilePath) }()

	conn := openCVConnection{}
	err = conn.connect(context.TODO(), mp4FilePath)
	require.NoError(t, err)

	err = conn.Close()
	require.NoError(t, err)
}

func TestOpenAndReadFromVideoStream(t *testing.T) {
	mp4FilePath, err := restoreMp4File()
	require.NoError(t, err)
	defer func() { os.Remove(mp4FilePath) }()

	conn := openCVConnection{}
	err = conn.connect(context.TODO(), mp4FilePath)
	require.NoError(t, err)

	frame := DefaultBackend().NewFrame()
	err = conn.Read(frame)
	require.NoError(t, err)
}

func TestOpenAndReadFromVideoStreamReadsToInternalFrameData(t *testing.T) {
	mp4FilePath, err := restoreMp4File()
	require.NoError(t, err)
	defer func() { os.Remove(mp4FilePath) }()

	conn := openCVConnection{}
	err = conn.connect(context.TODO(), mp4FilePath)
	require.NoError(t, err)

	frame := openCVFrame{
		mat: gocv.NewMat(),
	}
	defer frame.Close()

	assert.Zero(t, frame.mat.Total())
	err = conn.Read(frame)
	require.NoError(t, err)
	assert.Greater(t, frame.mat.Total(), 0)
}

type invalidFrame struct{}

func (frame invalidFrame) DataRef() interface{} {
	return nil
}

func (frame invalidFrame) Close() {}

func TestOpenAndReadWithIncorrectFrameDataReturnsError(t *testing.T) {
	mp4FilePath, err := restoreMp4File()
	require.NoError(t, err)
	defer func() { os.Remove(mp4FilePath) }()

	conn := openCVConnection{}
	err = conn.connect(context.TODO(), mp4FilePath)
	require.NoError(t, err)

	frame := invalidFrame{}
	err = conn.Read(frame)
	assert.EqualError(t, err, "must pass OpenCV frame to OpenCV connection read")
}

func TestOpenAndReadFailToReadFromConnectionReturnsError(t *testing.T) {
	resetReadFromVidCap := overloadReadFromVidCap(
		func(*gocv.VideoCapture, *gocv.Mat) bool {
			return false
		},
	)
	defer resetReadFromVidCap()

	mp4FilePath, err := restoreMp4File()
	require.NoError(t, err)
	defer func() { os.Remove(mp4FilePath) }()

	conn := openCVConnection{}
	err = conn.connect(context.TODO(), mp4FilePath)
	require.NoError(t, err)

	frame := openCVFrame{
		mat: gocv.NewMat(),
	}
	defer frame.Close()

	err = conn.Read(frame)
	assert.EqualError(t, err, "unable to read from video connection")
}
