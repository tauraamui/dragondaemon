package videostorage_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/video/videobackend"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
	"github.com/tauraamui/dragondaemon/pkg/video/videostorage"
)

func resolveTempDB(t *testing.T) (string, func()) {
	file, err := os.CreateTemp("", "dragonvideostore.*.db")
	if err != nil {
		t.Fatal(err)
	}

	tmpfilePath := file.Name()
	return tmpfilePath, func() { os.Remove(tmpfilePath) }
}

func TestNewStorageSuccess(t *testing.T) {
	tmpDB, cleanup := resolveTempDB(t)
	defer cleanup()

	is := is.New(t)

	s, err := videostorage.NewStorage(tmpDB)
	is.NoErr(err)
	is.True(s != nil)

	is.NoErr(s.Close())
}

func TestStoreFramesInStorageSuccess(t *testing.T) {
	tmpDB, cleanup := resolveTempDB(t)
	defer cleanup()

	is := is.New(t)

	s, err := videostorage.NewStorage(tmpDB) // using our custom timeseries file format to store video
	is.NoErr(err)
	is.True(s != nil)
	defer s.Close()

	frames := generateMockFrames(t, 60) // 1 second clip @ 60 fps will contain 60 frames

	timestamp := time.Date(2022, 4, 1, 10, 0, 0, 0, time.UTC).Unix()
	is.NoErr(s.SaveFrames(timestamp, frames))
}

// func TestStoreFramesLoadFramesInStorageSuccess(t *testing.T) {
// 	tmpDB, cleanup := resolveTempDB(t)
// 	defer cleanup()

// 	is := is.New(t)

// 	s, err := videostorage.NewStorage(tmpDB) // using our custom timeseries file format to store video
// 	is.NoErr(err)
// 	is.True(s != nil)
// 	defer s.Close()

// 	frames := generateMockFrames(t, 10)

// 	timestamp := time.Date(2022, 4, 1, 10, 0, 0, 0, time.UTC).Unix()
// 	is.NoErr(s.SaveFrames(timestamp, frames))
// }

func generateMockFrames(t *testing.T, c int) []videoframe.Frame {
	mock := videobackend.Mock()
	mockConn, err := mock.Connect(context.Background(), "addressdoesnotmatter")
	if err != nil {
		t.Fatal(err)
	}

	frames := []videoframe.Frame{}
	for i := 0; i < c; i++ {
		frame := mock.NewFrame()
		err := mockConn.Read(frame)
		if err != nil {
			t.Fatal(err)
			break
		}
		frames = append(frames, frame)
	}

	return frames
}
