package process

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tacusci/logging/v2"
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

func overloadErrorLog(overload func(string, ...interface{})) func() {
	logErrorRef := log.Error
	log.Error = overload
	return func() { log.Error = logErrorRef }
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
	errorLogs              []string
	resetErrorLogsOverload func()
}

func (suite *StreamAndPersistProcessesTestSuite) SetupSuite() {
	logging.CurrentLoggingLevel = logging.SilentLevel
	suite.backend = video.MockBackend()
	conn, err := camera.Connect("TestConn", "fake-addr", camera.Settings{}, suite.backend)
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

	resetLogError := overloadErrorLog(
		func(format string, a ...interface{}) {
			suite.errorLogs = append(suite.errorLogs, fmt.Sprintf(format, a...))
		},
	)
	suite.resetErrorLogsOverload = resetLogError
}

func (suite *StreamAndPersistProcessesTestSuite) TearDownTest() {
	suite.errorLogs = nil
	suite.debugLogs = nil
	suite.infoLogs = nil
	suite.resetInfoLogsOverload()
	suite.resetDebugLogsOverload()
	suite.resetErrorLogsOverload()
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

func (suite *StreamAndPersistProcessesTestSuite) TestStreamProcess() {
	frames := make(chan video.Frame)

	count := countFramesReadFromStreamProc(suite.conn, frames, 10)
	assert.Equal(suite.T(), 11, count)
}

func (suite *StreamAndPersistProcessesTestSuite) TestGenerateClipsProcess() {
	const FPS = 30
	const SPC = 2
	const expectedClipCount = 6

	count := countClipsCreatedByGenerateProc(suite.backend, FPS, SPC, expectedClipCount, defaultFrames)

	assert.GreaterOrEqual(suite.T(), count, expectedClipCount)
	assert.LessOrEqual(suite.T(), expectedClipCount, count+2)
}

func (suite *StreamAndPersistProcessesTestSuite) TestGenerateClipsProcessExtraFrames() {
	const FPS = 30
	const SPC = 2
	const expectedClipCount = 6

	frames := func(backend video.Backend, fps, spc, expectedCount int, frames chan video.Frame, done chan interface{}) {
		for i := 0; i < ((fps*spc)*expectedCount)+12; i++ {
			frames <- backend.NewFrame()
		}
		close(done)
	}

	count := countClipsCreatedByGenerateProc(suite.backend, FPS, SPC, expectedClipCount, frames)

	assert.Equal(suite.T(), expectedClipCount+1, count)
}

func (suite *StreamAndPersistProcessesTestSuite) TestGenerateClipsProcessMissingFrames() {
	const FPS = 30
	const SPC = 2
	const expectedClipCount = 6

	frames := func(backend video.Backend, fps, spc, expectedCount int, frames chan video.Frame, done chan interface{}) {
		for i := 0; i < ((fps*spc)*expectedCount)-69; i++ {
			frames <- backend.NewFrame()
		}
		close(done)
	}

	count := countClipsCreatedByGenerateProc(suite.backend, FPS, SPC, expectedClipCount, frames)

	assert.Equal(suite.T(), expectedClipCount-1, count)
}

type testVideoClip struct {
	onCloseCallback func()
}

func (clip testVideoClip) AppendFrame(video.Frame) {}

func (clip testVideoClip) FrameDimensions() (int, int) {
	return 100, 50
}

func (clip testVideoClip) FPS() int {
	return 30
}

func (clip testVideoClip) FileName() string {
	return "fake-video-clip.mp4"
}

func (clip testVideoClip) Close() {
	if clip.onCloseCallback != nil {
		clip.onCloseCallback()
	}
}

func (clip testVideoClip) GetFrames() []video.Frame {
	return nil
}

type testClipWriter struct {
	onWrite    func()
	writeError error
}

func (w testClipWriter) Write(video.Clip) error {
	if w.onWrite != nil {
		w.onWrite()
	}
	if w.writeError != nil {
		return w.writeError
	}
	return nil
}

func (suite *StreamAndPersistProcessesTestSuite) TestWriteClipsToDiskProcess() {
	const writeCount = 55
	const closeCount = writeCount

	clips := make(chan video.Clip)
	writeInvokedCount := 0
	writeClipsProcess := WriteClipsToDiskProcess(clips, testClipWriter{
		onWrite: func() { writeInvokedCount++ },
	})
	ctx, cancel := context.WithCancel(context.TODO())

	writeClipsProcess(ctx)

	closeInvokedCount := 0
	for i := 0; i < writeCount; i++ {
		clip := testVideoClip{
			onCloseCallback: func() { closeInvokedCount++ },
		}
		clips <- clip
	}

	cancel()

	assert.Equal(suite.T(), writeCount, writeInvokedCount)
	assert.Equal(suite.T(), closeCount, closeInvokedCount)
}

func (suite *StreamAndPersistProcessesTestSuite) TestWriteClipsToDiskProcessLogsWriteFailErrors() {
	clips := make(chan video.Clip)
	writeClipsProcess := WriteClipsToDiskProcess(clips, testClipWriter{
		writeError: errors.New("clip write test error"),
	})
	ctx, cancel := context.WithCancel(context.TODO())

	writeClipsProcess(ctx)

	for i := 0; i < 3; i++ {
		clips <- testVideoClip{}
	}

	cancel()

	assert.Len(suite.T(), suite.errorLogs, 3)
	assert.ElementsMatch(suite.T(), suite.errorLogs, []string{
		"Unable to write clip to disk: clip write test error",
		"Unable to write clip to disk: clip write test error",
		"Unable to write clip to disk: clip write test error",
	})
}

func defaultFrames(
	backend video.Backend,
	fps, spc, expectedCount int,
	frames chan video.Frame, done chan interface{},
) {
	for i := 0; i < (fps*spc)*expectedCount; i++ {
		frames <- backend.NewFrame()
	}
	close(done)
}

type counter struct {
	sync.Mutex
	count int
}

func (c *counter) value() int {
	c.Lock()
	defer c.Unlock()
	return c.count
}

func (c *counter) incr() {
	c.Lock()
	defer c.Unlock()
	c.count++
}

func countFramesReadFromStreamProc(conn camera.Connection, frames chan video.Frame, targetToSend int) int {
	runStreamProcess := StreamProcess(conn, frames)
	ctx, cancel := context.WithCancel(context.TODO())
	runStreamProcess(ctx)

	c := counter{}
	go func(cancel context.CancelFunc, count *counter) {
		for {
			if count.value() >= targetToSend {
				cancel()
				break
			}
		}
	}(cancel, &c)

procLoop:
	for {
		select {
		case <-ctx.Done():
			break procLoop
		default:
			f := <-frames
			f.Close()
			c.incr()
		}
	}
	return c.value()
}

func countClipsCreatedByGenerateProc(
	backend video.Backend,
	fps, spc, expectedCount int,
	frameMaker func(video.Backend, int, int, int, chan video.Frame, chan interface{}),
) int {
	frames := make(chan video.Frame)

	doneCreatingFrames := make(chan interface{})
	go frameMaker(backend, fps, spc, expectedCount, frames, doneCreatingFrames)

	countingCtx, cancelClipCount := context.WithCancel(context.TODO())
	clips := make(chan video.Clip)
	c := counter{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup, count *counter, clips chan video.Clip, stop context.Context) {
		defer wg.Done()
	procLoop:
		for {
			select {
			case <-stop.Done():
				break procLoop
			default:
				clip := <-clips
				c.incr()
				clip.Close()
			}
		}
	}(&wg, &c, clips, countingCtx)

	procCtx, cancelProc := context.WithCancel(context.TODO())
	proc := GenerateClipsProcess(frames, clips, "", fps, spc)
	go proc(procCtx)

	<-doneCreatingFrames
	cancelProc()
	cancelClipCount()

	wg.Wait()
	return c.value()
}
