package videobackend_test

import (
	"context"
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/video/videobackend"
)

func TestRTSPClientConn(t *testing.T) {
	is := is.New(t)
	client := videobackend.NewRTSPClient()
	is.NoErr(client.Connect(context.Background(), "rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mov"))
}
