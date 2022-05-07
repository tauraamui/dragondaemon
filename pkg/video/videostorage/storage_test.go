package videostorage_test

import (
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/video/videostorage"
)

func TestLoadStorageSuccess(t *testing.T) {
	is := is.New(t)

	s, err := videostorage.NewStorage()
	is.NoErr(err)
	is.True(s != nil)

	is.NoErr(s.SaveFrames(time.Now().Unix(), nil))
}
