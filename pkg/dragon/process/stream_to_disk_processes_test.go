package process

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/internal/videotest"
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

func overloadInfoLog(overload func(string, ...interface{})) func() {
	logInfoRef := log.Debug
	log.Info = overload
	return func() { log.Info = logInfoRef }
}

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
	infoLogs               []string
	resetInfoLogsOverload  func()
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
	suite.infoLogs = []string{}
	resetLogInfo := overloadInfoLog(
		func(format string, a ...interface{}) {
			suite.infoLogs = append(suite.infoLogs, fmt.Sprintf(format, a...))
		},
	)
	suite.resetInfoLogsOverload = resetLogInfo

	suite.debugLogs = []string{}
	resetLogDebug := overloadDebugLog(
		func(format string, a ...interface{}) {
			suite.debugLogs = append(suite.debugLogs, fmt.Sprintf(format, a...))
		},
	)
	suite.resetDebugLogsOverload = resetLogDebug
}

func (suite *StreamAndPersistProcessesTestSuite) TearDownTest() {
	suite.resetInfoLogsOverload()
	suite.resetDebugLogsOverload()
}

func TestStreamAndPersistProcessTestSuite(t *testing.T) {
	suite.Run(t, &StreamAndPersistProcessesTestSuite{})
}

// Probably replace this test with a unit test, and then a e2e test for the whole process
func (suite *StreamAndPersistProcessesTestSuite) TestStreamProcessWithRealImpl() {
	frames := make(chan video.Frame)
	runStreamProcess := StreamProcess(suite.conn, frames)
	ctx, cancel := context.WithCancel(context.TODO())
	runStreamProcess(ctx)
	time.Sleep(5 * time.Millisecond)
	cancel()
	assert.Contains(suite.T(), suite.infoLogs,
		"Streaming video from camera [TestConn]",
	)
	assert.Contains(suite.T(), suite.debugLogs,
		"Reading frame from vid stream for camera [TestConn]",
		"Buffer full...",
	)
}

func (suite *StreamAndPersistProcessesTestSuite) TestGenerateClipsProcess() {
	const FPS = 30
	const SPC = 2
	const expectedClipCount = 6

	count := countClipsCreatedByGenerateProc(FPS, SPC, expectedClipCount)

	assert.Equal(suite.T(), expectedClipCount, count)
}

func countClipsCreatedByGenerateProc(fps, spc, expectedCount int) int {
	var backend = video.DefaultBackend()

	frames := make(chan video.Frame)

	doneCreatingFrames := make(chan interface{})
	go func(frames chan video.Frame, done chan interface{}) {
		for i := 0; i < (fps*spc)*expectedCount; i++ {
			frames <- backend.NewFrame()
		}
		close(done)
	}(frames, doneCreatingFrames)

	countingCtx, cancelClipCount := context.WithCancel(context.TODO())
	clips := make(chan video.Clip)
	count := 0
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup, count *int, clips chan video.Clip, stop context.Context) {
		defer wg.Done()
	procLoop:
		for {
			select {
			case <-stop.Done():
				break procLoop
			default:
				clip := <-clips
				*count++
				clip.Close()
			}
		}
	}(&wg, &count, clips, countingCtx)

	procCtx, cancelProc := context.WithCancel(context.TODO())
	proc := GenerateClipsProcess(frames, clips, fps, spc)
	go proc(procCtx)

	<-doneCreatingFrames
	cancelProc()
	cancelClipCount()

	wg.Wait()
	return count
}
