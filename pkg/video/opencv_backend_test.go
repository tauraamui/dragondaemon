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

func TestOpenVideoStream(t *testing.T) {
	mp4Dir := os.TempDir()
	err := RestoreAsset(mp4Dir, "small.mp4")
	require.NoError(t, err)

	mp4FilePath := filepath.Join(mp4Dir, "small.mp4")
	defer os.Remove(mp4FilePath)

	conn := openCVConnection{}
	err = conn.connect(context.TODO(), mp4FilePath)
	require.NoError(t, err)
}
