package dragon_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tauraamui/dragondaemon/internal/videotest"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/dragon"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type ServerProcessTestSuite struct {
	suite.Suite
	mp4FilePath string
	server      dragon.Server
}

func (suite *ServerProcessTestSuite) SetupSuite() {
	mp4FilePath, err := videotest.RestoreMp4File()
	require.NoError(suite.T(), err)
	suite.mp4FilePath = mp4FilePath
}

func (suite *ServerProcessTestSuite) TeardownSuite() {
	require.NoError(suite.T(), os.Remove(suite.mp4FilePath))
}

func (suite *ServerProcessTestSuite) SetupTest() {
	suite.server = dragon.NewServer(testConfigResolver{
		resolveConfigs: func() configdef.Values {
			return configdef.Values{
				Cameras: []configdef.Camera{
					{Title: "TestConn", Address: suite.mp4FilePath},
				},
			}
		},
	}, video.DefaultBackend())
}

func (suite *ServerProcessTestSuite) TestRunProcesses() {
	require.NoError(suite.T(), suite.server.LoadConfiguration())
	require.Len(suite.T(), suite.server.Connect(), 0)
	suite.server.RunProcesses()
	<-suite.server.Shutdown()
}

func TestServerProcessTestSuite(t *testing.T) {
	suite.Run(t, &ServerProcessTestSuite{})
}

// func TestServerRunProcesses(t *testing.T) {
// 	mp4FilePath, err := videotest.RestoreMp4File()
// 	require.NoError(t, err)
// 	defer func() { os.Remove(mp4FilePath) }()

// 	s := dragon.NewServer(testConfigResolver{
// 		resolveConfigs: func() configdef.Values {
// 			return configdef.Values{
// 				Cameras: []configdef.Camera{
// 					{Title: "TestConn", Address: mp4FilePath},
// 				},
// 			}
// 		},
// 	}, video.DefaultBackend())
// 	require.NoError(t, s.LoadConfiguration())
// 	require.Len(t, s.Connect(), 0)
// 	s.RunProcesses()
// 	<-s.Shutdown()
// }
