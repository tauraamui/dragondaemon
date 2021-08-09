package process_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/internal/videotest"
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/dragon/process"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

func overloadDebugLog(overload func(string, ...interface{})) func() {
	logDebugRef := log.Debug
	log.Debug = overload
	return func() { log.Info = logDebugRef }
}

type StreamAndPersistProcessesTestSuite struct {
	suite.Suite
	mp4FilePath            string
	backend                video.Backend
	conn                   camera.Connection
	debugLogs              []string
	resetDebugLogsOverload func()
}

func (suite *StreamAndPersistProcessesTestSuite) SetupSuite() {
	logging.CurrentLoggingLevel = logging.DebugLevel
	mp4FilePath, err := videotest.RestoreMp4File()
	require.NoError(suite.T(), err)
	suite.mp4FilePath = mp4FilePath

	suite.backend = video.DefaultBackend()
	conn, err := camera.Connect("TestConn", mp4FilePath, camera.Settings{}, suite.backend)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), conn)
	suite.conn = conn
}

func (suite *StreamAndPersistProcessesTestSuite) TearDownSuite() {
	logging.CurrentLoggingLevel = logging.WarnLevel
	file, err := os.Open(suite.mp4FilePath)
	if err == nil {
		os.Remove(file.Name())
	}
	file.Close()
}

func (suite *StreamAndPersistProcessesTestSuite) SetupTest() {
	suite.debugLogs = []string{}
	resetLogDebug := overloadDebugLog(
		func(format string, a ...interface{}) {
			suite.debugLogs = append(suite.debugLogs, fmt.Sprintf(format, a...))
		},
	)
	suite.resetDebugLogsOverload = resetLogDebug
}

func (suite *StreamAndPersistProcessesTestSuite) TearDownTest() {
	suite.resetDebugLogsOverload()
}

func (suite *StreamAndPersistProcessesTestSuite) TestStreamProcess() {
	frames := make(chan video.Frame)
	runStreamProcess := process.StreamProcess(suite.conn, frames)
	ctx, cancel := context.WithCancel(context.TODO())
	runStreamProcess(ctx)
	time.Sleep(5 * time.Millisecond)
	cancel()
	assert.Contains(suite.T(), suite.debugLogs,
		"Reading frame from vid stream for camera [TestConn]",
		"Buffer full...",
	)
}

func TestStreamAndPersistProcessTestSuite(t *testing.T) {
	suite.Run(t, &StreamAndPersistProcessesTestSuite{})
}
