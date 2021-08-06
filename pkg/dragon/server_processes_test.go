package dragon_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tauraamui/dragondaemon/internal/videotest"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/dragon"
)

func TestServerRunProcesses(t *testing.T) {
	mp4FilePath, err := videotest.RestoreMp4File()
	require.NoError(t, err)
	defer func() { os.Remove(mp4FilePath) }()

	s := dragon.NewServer(testConfigResolver{
		resolveConfigs: func() configdef.Values {
			return configdef.Values{
				Cameras: []configdef.Camera{},
			}
		},
	}, testVideoBackend{})
	s.RunProcesses()
	<-s.Shutdown()
}
