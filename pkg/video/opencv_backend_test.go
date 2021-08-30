package video

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

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

func overloadOpenVideoWriter(overload func(filename, codec string, fps float64, width, height int, isColor bool) (*gocv.VideoWriter, error)) func() {
	openVidWriterRef := openVideoWriter
	openVideoWriter = overload
	return func() { openVideoWriter = openVidWriterRef }
}

func setupOSFSForTesting() (string, error) {
	fs = afero.NewOsFs()
	rootDir, err := videotest.MakeRootPath(fs)
	if err != nil {
		return "", err
	}
	return rootDir, nil
}

func overloadTimestamp(fixed time.Time) func() {
	return overloadTimestampFunc(func() time.Time { return fixed })
}

func overloadTimestampFunc(overload func() time.Time) func() {
	TimestampRef := Timestamp
	Timestamp = overload
	return func() { Timestamp = TimestampRef }
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
	dimensions := frame.Dimensions()
	assert.Equal(t, dimensions.W, 560)
	assert.Equal(t, dimensions.H, 320)
	// make sure as much as possible that the impl
	// isn't just writing random junk to the frame
	assert.Equal(t, frame.mat.ToBytes()[:10], []byte{
		0xe, 0x27, 0x48, 0xe, 0x27, 0x48, 0xe, 0x27, 0x48, 0xe,
	})
}

func makeClips(rootPath string, seconds, fps, count int) ([]Clip, error) {
	clips := []Clip{}
	for i := 0; i < count; i++ {
		clip, err := makeClip(rootPath, 3, 10)
		if err != nil {

			return nil, err
		}
		clips = append(clips, clip)
	}
	return clips, nil
}

func makeClip(rootPath string, seconds, fps int) (Clip, error) {
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

	clip := NewClip(fmt.Sprintf("/%s/clips/TestCam", rootPath), fps)
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

func (frame invalidFrame) Dimensions() FrameDimension {
	return FrameDimension{W: 100, H: 50}
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
	resetTimestamp := overloadTimestamp(time.Unix(1630184250, 0).UTC())
	defer resetTimestamp()

	clip, err := makeClip("testroot", 3, 10)
	require.NoError(t, err)

	var passedFilename string
	var passedCodec string
	var passedFPS float64
	var passedWidth, passedHeight int
	var passedIsColor bool
	resetOpenVidWriter := overloadOpenVideoWriter(func(
		filename, codec string, fps float64, width, height int, isColor bool,
	) (*gocv.VideoWriter, error) {
		passedFilename = filename
		passedCodec = codec
		passedFPS = fps
		passedWidth = width
		passedHeight = height
		passedIsColor = isColor
		return &gocv.VideoWriter{}, nil
	})
	defer resetOpenVidWriter()

	writer := openCVClipWriter{}
	err = writer.init(clip)
	assert.NoError(t, err)

	assert.Equal(t, "/testroot/clips/TestCam/2021-08-28/2021-08-28 20.57.30.mp4", passedFilename)
	assert.Equal(t, codec, passedCodec)
	assert.EqualValues(t, 10, passedFPS)
	assert.Equal(t, 560, passedWidth)
	assert.Equal(t, 320, passedHeight)
	assert.True(t, passedIsColor)
}

func TestClipWriterWrite(t *testing.T) {
	resetTimestamp := overloadTimestamp(time.Unix(1630184250, 0).UTC())
	defer resetTimestamp()

	pathRoot, err := setupOSFSForTesting()
	require.NoError(t, err)
	defer func() {
		os.RemoveAll(pathRoot)
	}()

	clip, err := makeClip(pathRoot, 3, 10)
	require.NoError(t, err)
	require.NotNil(t, clip)

	backend := openCVBackend{}
	writer := backend.NewWriter()
	require.NotNil(t, writer)
	err = writer.Write(clip)
	assert.NoError(t, err)

	_, err = fs.Stat(fmt.Sprintf("/%s/clips/TestCam/2021-08-28/2021-08-28 20.57.30.mp4", pathRoot))
	assert.NoError(t, err)
}

func TestClipWriterWriteMultipleClips(t *testing.T) {
	const seconds = 2
	const fps = 10
	const clipCount = 8

	var accessCount int64 = 0
	resetTimestampFunc := overloadTimestampFunc(func() time.Time {
		accessCount += seconds
		return time.Unix(1630184250+accessCount, 0).UTC()
	})
	defer resetTimestampFunc()

	backend := openCVBackend{}
	writer := backend.NewWriter()
	require.NotNil(t, writer)

	pathRoot, err := setupOSFSForTesting()
	require.NoError(t, err)
	defer func() {
		println("Removing all under [%s]", pathRoot)
	}()

	clips, err := makeClips(pathRoot, seconds, fps, clipCount)
	require.NoError(t, err)

	for _, clip := range clips {
		err = writer.Write(clip)
		require.NoError(t, err)
		clip.Close()
	}

	expectedClipFiles := []string{
		// first clip in list should not exist
		pathRoot + "/clips/TestCam/2021-08-28/2021-08-28 20.57.30.mp4",
		pathRoot + "/clips/TestCam/2021-08-28/2021-08-28 20.57.32.mp4",
		pathRoot + "/clips/TestCam/2021-08-28/2021-08-28 20.57.34.mp4",
		pathRoot + "/clips/TestCam/2021-08-28/2021-08-28 20.57.36.mp4",
		pathRoot + "/clips/TestCam/2021-08-28/2021-08-28 20.57.38.mp4",
		pathRoot + "/clips/TestCam/2021-08-28/2021-08-28 20.57.40.mp4",
		pathRoot + "/clips/TestCam/2021-08-28/2021-08-28 20.57.42.mp4",
		pathRoot + "/clips/TestCam/2021-08-28/2021-08-28 20.57.44.mp4",
		pathRoot + "/clips/TestCam/2021-08-28/2021-08-28 20.57.46.mp4",
		// last clip in list should not exist
		pathRoot + "/clips/TestCam/2021-08-28/2021-08-28 20.57.48.mp4",
	}

	for i, path := range expectedClipFiles {
		clip, err := fs.Stat(path)
		if i > 0 && i < clipCount+1 {
			assert.False(t, clip.IsDir())
			assert.NoError(t, err)
			continue
		}
		assert.ErrorIs(t, err, os.ErrNotExist)
	}
}
