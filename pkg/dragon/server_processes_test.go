package dragon_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/dragon"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type ServerProcessTestSuite struct {
	suite.Suite
	server                dragon.Server
	infoLogs              []string
	resetInfoLogsOverload func()
}

func (suite *ServerProcessTestSuite) SetupTest() {
	logging.CurrentLoggingLevel = logging.SilentLevel
	suite.server = dragon.NewServer(testConfigResolver{
		resolveConfigs: func() configdef.Values {
			return configdef.Values{
				Cameras: []configdef.Camera{
					{Title: "TestConn", Address: "fake-conn-addr"},
				},
			}
		},
	}, video.MockBackend())

	suite.infoLogs = []string{}
	resetLogInfo := overloadInfoLog(
		func(format string, a ...interface{}) {
			suite.infoLogs = append(suite.infoLogs, fmt.Sprintf(format, a...))
		},
	)
	suite.resetInfoLogsOverload = resetLogInfo
}

func (suite *ServerProcessTestSuite) TearDownTest() {
	logging.CurrentLoggingLevel = logging.WarnLevel
	suite.resetInfoLogsOverload()
}

func (suite *ServerProcessTestSuite) TestRunProcesses() {
	require.NoError(suite.T(), suite.server.LoadConfiguration())
	require.Len(suite.T(), suite.server.Connect(), 0)
	suite.server.SetupProcesses()
	suite.server.RunProcesses()
	time.Sleep(1 * time.Millisecond)
	<-suite.server.Shutdown()
	assert.Subset(suite.T(), suite.infoLogs, []string{
		"Connecting to camera: [TestConn@fake-conn-addr]...",
		"Connected successfully to camera: [TestConn]",
		"Streaming video from camera [TestConn]",
		"Closing camera [TestConn] video stream...",
	})
}

func TestServerProcessTestSuite(t *testing.T) {
	suite.Run(t, &ServerProcessTestSuite{})
}
