package dragon_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/dragon"
	"github.com/tauraamui/dragondaemon/pkg/video"
	"github.com/tauraamui/dragondaemon/pkg/xis"
)

type ServerProcessTestSuite struct {
	suite.Suite
	server                *dragon.Server
	infoLogs              []string
	resetInfoLogsOverload func()
}

func (suite *ServerProcessTestSuite) SetupTest() {
	is := is.New(suite.T())
	logging.CurrentLoggingLevel = logging.SilentLevel
	svr, err := dragon.NewServer(testConfigResolver{
		resolveConfigs: func() configdef.Values {
			return configdef.Values{
				Cameras: []configdef.Camera{
					{Title: "TestConn", Address: "fake-conn-addr"},
				},
			}
		},
	}, video.MockBackend())
	is.True(svr != nil)
	is.NoErr(err)

	suite.server = svr

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
	require.Len(suite.T(), suite.server.Connect(), 0)
	suite.server.SetupProcesses()
	suite.server.RunProcesses()
	time.Sleep(1 * time.Millisecond)
	<-suite.server.Shutdown()
	xis := xis.New(is.New(suite.T()))
	xis.Subset(suite.infoLogs, []string{
		"Connecting to camera: [TestConn@fake-conn-addr]...",
		"Connected successfully to camera: [TestConn]",
		"Streaming video from camera [TestConn]",
		"Closing camera [TestConn] video stream...",
	})
}

func TestServerProcessTestSuite(t *testing.T) {
	suite.Run(t, &ServerProcessTestSuite{})
}
