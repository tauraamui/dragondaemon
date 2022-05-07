package rtsp_test

import (
	"context"
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/rtsp"
)

func TestRTSPClientConn(t *testing.T) {
	is := is.New(t)
	client, err := rtsp.NewClient("rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mov")
	is.NoErr(err)
	is.NoErr(client.Connect(context.Background()))
	is.NoErr(client.Options())
	is.NoErr(client.Close())
}
