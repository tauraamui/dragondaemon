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

func restoreMp4File() (string, error) {
	mp4Dir := os.TempDir()
	err := RestoreAsset(mp4Dir, "small.mp4")
	if err != nil {
		return "", err
	}

	return filepath.Join(mp4Dir, "small.mp4"), nil
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

func TestOpenAndReadWithIncorrectFrameDataReturnsError(t *testing.T) {

}
