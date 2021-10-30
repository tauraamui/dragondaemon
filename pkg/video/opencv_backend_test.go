package video

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tauraamui/dragondaemon/internal/videotest"
	"github.com/tauraamui/xerror"
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
	is := is.New(t)
	mp4FilePath, err := videotest.RestoreMp4File()
	require.NoError(t, err)
	defer func() { os.Remove(mp4FilePath) }()

	backend := openCVBackend{}
	conn, err := backend.Connect(context.TODO(), mp4FilePath)
	is.NoErr(err)
	is.True(conn != nil)
	err = conn.Close()
	is.True(err == nil)
}

func TestBackendConnectWithImmediateCancelInvoke(t *testing.T) {
	is := is.New(t)
	mp4FilePath, err := videotest.RestoreMp4File()
	is.NoErr(err)
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
	is.Equal(connErr.Error(), "connection cancelled")
}

func TestConnectWithImmediateCancelInvoke(t *testing.T) {
	is := is.New(t)
	mp4FilePath, err := videotest.RestoreMp4File()
	is.NoErr(err)
	defer func() { os.Remove(mp4FilePath) }()

	conn := openCVConnection{}

	ctx, cancel := context.WithCancel(context.TODO())
	errChan := make(chan error)
	go func(ctx context.Context) {
		errChan <- conn.connect(ctx, mp4FilePath)
	}(ctx)
	cancel()

	connErr := <-errChan
	is.Equal(connErr.Error(), "connection cancelled")
}

func TestOpenVideoStreamInvokesOpenVideoCapture(t *testing.T) {
	is := is.New(t)
	resetOpenVidCap := overloadOpenVidCap(
		func(addr string) (*gocv.VideoCapture, error) {
			return nil, xerror.New("test connect error")
		},
	)
	defer resetOpenVidCap()

	conn := openCVConnection{}
	is.Equal(conn.connect(context.TODO(), "TestAddr").Error(), "test connect error")
}

func TestOpenAndCloseVideoStream(t *testing.T) {
	is := is.New(t)
	mp4FilePath, err := videotest.RestoreMp4File()
	is.NoErr(err)
	defer func() { os.Remove(mp4FilePath) }()

	conn := openCVConnection{}
	is.NoErr(conn.connect(context.TODO(), mp4FilePath))
	is.NoErr(err)

	err = conn.Close()
	is.NoErr(err)
}

func TestOpenAndReadFromVideoStream(t *testing.T) {
	is := is.New(t)
	mp4FilePath, err := videotest.RestoreMp4File()
	is.NoErr(err)
	defer func() { os.Remove(mp4FilePath) }()

	conn := openCVConnection{}
	is.NoErr(conn.connect(context.TODO(), mp4FilePath))

	frame := DefaultBackend().NewFrame()
	is.NoErr(conn.Read(frame))
}

func TestOpenAndReadFromVideoStreamReadsToInternalFrameData(t *testing.T) {
	is := is.New(t)
	mp4FilePath, err := videotest.RestoreMp4File()
	is.NoErr(err)
	defer func() { os.Remove(mp4FilePath) }()

	conn := openCVConnection{}
	is.NoErr(conn.connect(context.TODO(), mp4FilePath))

	frame := &openCVFrame{
		mat: gocv.NewMat(),
	}
	defer frame.Close()

	is.Equal(frame.mat.Total(), 0)
	is.NoErr(conn.Read(frame))
	is.True(frame.mat.Total() > 0)
	dimensions := frame.Dimensions()
	is.Equal(dimensions.W, 560)
	is.Equal(dimensions.H, 320)
	// make sure as much as possible that the impl
	// isn't just writing random junk to the frame
	is.Equal(frame.mat.ToBytes()[:10], []byte{
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
	is := is.New(t)
	mp4FilePath, err := videotest.RestoreMp4File()
	is.NoErr(err)
	defer func() { os.Remove(mp4FilePath) }()

	conn := openCVConnection{}
	is.NoErr(conn.connect(context.TODO(), mp4FilePath))

	frame := invalidFrame{}
	is.Equal(conn.Read(frame).Error(), "must pass OpenCV frame to OpenCV connection read")
}

func TestOpenAndReadFailToReadFromConnectionReturnsError(t *testing.T) {
	is := is.New(t)
	resetReadFromVidCap := overloadReadFromVidCap(
		func(*gocv.VideoCapture, *gocv.Mat) bool {
			return false
		},
	)
	defer resetReadFromVidCap()

	mp4FilePath, err := videotest.RestoreMp4File()
	is.NoErr(err)
	defer func() { os.Remove(mp4FilePath) }()

	conn := openCVConnection{}
	is.NoErr(conn.connect(context.TODO(), mp4FilePath))

	frame := &openCVFrame{
		mat: gocv.NewMat(),
	}
	defer frame.Close()

	is.Equal(conn.Read(frame).Error(), "unable to read from video connection")
}

func TestNewWriterReturnsNonNilInstance(t *testing.T) {
	is := is.New(t)
	backend := openCVBackend{}
	writer := backend.NewWriter()
	is.True(writer != nil)
}

func TestClipWriterInit(t *testing.T) {
	is := is.New(t)
	resetTimestamp := overloadTimestamp(time.Unix(1630184250, 0).UTC())
	defer resetTimestamp()

	clip, err := makeClip("testroot", 3, 10)
	is.NoErr(err)

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
	is.NoErr(writer.init(clip))

	is.Equal("/testroot/clips/TestCam/2021-08-28/2021-08-28 20.57.30.mp4", passedFilename)
	is.Equal(codec, passedCodec)
	is.Equal(10, int(passedFPS))
	is.Equal(560, passedWidth)
	is.Equal(320, passedHeight)
	is.True(passedIsColor)
}

func TestClipWriterWrite(t *testing.T) {
	is := is.New(t)
	resetTimestamp := overloadTimestamp(time.Unix(1630184250, 0).UTC())
	defer resetTimestamp()

	pathRoot, err := setupOSFSForTesting()
	require.NoError(t, err)
	defer func() {
		os.RemoveAll(pathRoot)
	}()

	clip, err := makeClip(pathRoot, 3, 10)
	is.NoErr(err)
	is.True(clip != nil)

	backend := openCVBackend{}
	writer := backend.NewWriter()
	is.True(writer != nil)
	err = writer.Write(clip)
	is.NoErr(err)

	_, err = fs.Stat(fmt.Sprintf("/%s/clips/TestCam/2021-08-28/2021-08-28 20.57.30.mp4", pathRoot))
	is.NoErr(err)
}

func TestClipWriterWriteMultipleClips(t *testing.T) {
	is := is.New(t)
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
	is.True(writer != nil)

	pathRoot, err := setupOSFSForTesting()
	is.NoErr(err)
	defer func() {
		println("Removing all under [%s]", pathRoot)
	}()

	clips, err := makeClips(pathRoot, seconds, fps, clipCount)
	is.NoErr(err)

	for _, clip := range clips {
		err = writer.Write(clip)
		is.NoErr(err)
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
			is.True(clip.IsDir() == false)
			is.NoErr(err)
			continue
		}
		is.True(errors.Is(err, os.ErrNotExist))
		assert.ErrorIs(t, err, os.ErrNotExist)
	}
}
