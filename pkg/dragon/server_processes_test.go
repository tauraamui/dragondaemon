package dragon_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tauraamui/dragondaemon/internal/videotest"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/dragon"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type ServerProcessTestSuite struct {
	suite.Suite
	mp4FilePath           string
	server                dragon.Server
	infoLogs              []string
	resetInfoLogsOverload func()
}

func (suite *ServerProcessTestSuite) SetupSuite() {
	mp4FilePath, err := videotest.RestoreMp4File()
	require.NoError(suite.T(), err)
	suite.mp4FilePath = mp4FilePath
}

func (suite *ServerProcessTestSuite) TearDownSuite() {
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

	suite.infoLogs = []string{}
	resetLogInfo := overloadInfoLog(
		func(format string, a ...interface{}) {
			suite.infoLogs = append(suite.infoLogs, fmt.Sprintf(format, a...))
		},
	)
	suite.resetInfoLogsOverload = resetLogInfo
}

func (suite *ServerProcessTestSuite) TearDownTest() {
	suite.resetInfoLogsOverload()
}

func (suite *ServerProcessTestSuite) TestRunProcesses() {
	require.NoError(suite.T(), suite.server.LoadConfiguration())
	require.Len(suite.T(), suite.server.Connect(), 0)
	suite.server.RunProcesses()
	<-suite.server.Shutdown()
	assert.Equal(suite.T(), []string{
		"Connecting to camera: [TestConn]...",
		"Connected successfully to camera: [TestConn]",
		"Streaming video from camera [TestConn]",
		"Stopping generating clips from [TestConn] video stream...",
		"Closing camera [TestConn] video stream...",
	}, suite.infoLogs)
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
