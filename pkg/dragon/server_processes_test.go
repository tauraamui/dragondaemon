package dragon_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tauraamui/dragondaemon/internal/videotest"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/dragon"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

func TestServerRunProcesses(t *testing.T) {
	mp4FilePath, err := videotest.RestoreMp4File()
	require.NoError(t, err)
	defer func() { os.Remove(mp4FilePath) }()

	s := dragon.NewServer(testConfigResolver{
		resolveConfigs: func() configdef.Values {
			return configdef.Values{
				Cameras: []configdef.Camera{
					{Title: "TestConn", Address: mp4FilePath},
				},
			}
		},
	}, video.DefaultBackend())
	require.NoError(t, s.LoadConfiguration())
	require.Len(t, s.Connect(), 0)
	s.RunProcesses()
	<-s.Shutdown()
}
