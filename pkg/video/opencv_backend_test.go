package video

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tauraamui/dragondaemon/internal/videotest"
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

func overloadFSIntoInMem() func() {
	fs = afero.NewMemMapFs()
	return func() { fs = afero.NewOsFs() }
}

func TestBackendConnect(t *testing.T) {
	mp4FilePath, err := videotest.RestoreMp4File()
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
	mp4FilePath, err := videotest.RestoreMp4File()
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
	mp4FilePath, err := videotest.RestoreMp4File()
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
	mp4FilePath, err := videotest.RestoreMp4File()
	require.NoError(t, err)
	defer func() { os.Remove(mp4FilePath) }()

	conn := openCVConnection{}
	err = conn.connect(context.TODO(), mp4FilePath)
	require.NoError(t, err)

	err = conn.Close()
	require.NoError(t, err)
}

func TestOpenAndReadFromVideoStream(t *testing.T) {
	mp4FilePath, err := videotest.RestoreMp4File()
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
	mp4FilePath, err := videotest.RestoreMp4File()
	require.NoError(t, err)
	defer func() { os.Remove(mp4FilePath) }()

	conn := openCVConnection{}
	err = conn.connect(context.TODO(), mp4FilePath)
	require.NoError(t, err)

	frame := &openCVFrame{
		mat: gocv.NewMat(),
	}
	defer frame.Close()

	assert.Zero(t, frame.mat.Total())
	err = conn.Read(frame)
	require.NoError(t, err)
	assert.Greater(t, frame.mat.Total(), 0)
	frameWidth, frameHeight := frame.Dimensions()
	assert.Equal(t, frameWidth, 560)
	assert.Equal(t, frameHeight, 320)
	// make sure as much as possible that the impl
	// isn't just writing random junk to the frame
	assert.Equal(t, frame.mat.ToBytes()[:10], []byte{
		0xe, 0x27, 0x48, 0xe, 0x27, 0x48, 0xe, 0x27, 0x48, 0xe,
	})
}

func makeClip(seconds, fps int) (Clip, error) {
	mp4FilePath, err := videotest.RestoreMp4File()
	if err != nil {
		return nil, err
	}
	defer func() { os.Remove(mp4FilePath) }()

	conn := openCVConnection{}
	err = conn.connect(context.TODO(), mp4FilePath)
	if err != nil {
		return nil, err
	}

	clip := NewClip("/testroot/clips/TestCam", fps)
	for i := 0; i < fps*seconds; i++ {
		f := &openCVFrame{
			mat: gocv.NewMat(),
		}
		err = conn.Read(f)
		if err != nil {
			f.Close()
			break
		}
		clip.AppendFrame(f)
	}

	if err != nil {
		return nil, err
	}

	return clip, err
}

type invalidFrame struct{}

func (frame invalidFrame) DataRef() interface{} {
	return nil
}

func (frame invalidFrame) Dimensions() (int, int) {
	return 100, 50
}

func (frame invalidFrame) Close() {}

func TestOpenAndReadWithIncorrectFrameDataReturnsError(t *testing.T) {
	mp4FilePath, err := videotest.RestoreMp4File()
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

	mp4FilePath, err := videotest.RestoreMp4File()
	require.NoError(t, err)
	defer func() { os.Remove(mp4FilePath) }()

	conn := openCVConnection{}
	err = conn.connect(context.TODO(), mp4FilePath)
	require.NoError(t, err)

	frame := &openCVFrame{
		mat: gocv.NewMat(),
	}
	defer frame.Close()

	err = conn.Read(frame)
	assert.EqualError(t, err, "unable to read from video connection")
}

func TestNewWriterReturnsNonNilInstance(t *testing.T) {
	backend := openCVBackend{}
	writer := backend.NewWriter()
	assert.NotNil(t, writer)
}

func TestClipWriterInit(t *testing.T) {
	overloadFSIntoInMem()
	defer func() { fs.RemoveAll("/") }()

	clip, err := makeClip(3, 10)
	require.NoError(t, err)

	writer := openCVClipWriter{}
	err = writer.init(clip)
	assert.NoError(t, err)
}
