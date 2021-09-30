package process_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/stretchr/testify/suite"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/dragon/process"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

func overloadErrorLog(overload func(string, ...interface{})) func() {
	logErrorRef := log.Error
	log.Error = overload
	return func() { log.Error = logErrorRef }
}

type StreamConnProcessTestSuite struct {
	suite.Suite
	resetErrorLogsOverload func()
	errorLogs              []string
}

func (suite *StreamConnProcessTestSuite) SetupSuite() {
	logging.CurrentLoggingLevel = logging.SilentLevel
}

func (suite *StreamConnProcessTestSuite) TearDownSuite() {
	logging.CurrentLoggingLevel = logging.WarnLevel
}

func (suite *StreamConnProcessTestSuite) SetupTest() {
	resetLogError := overloadErrorLog(
		func(format string, a ...interface{}) {
			suite.errorLogs = append(suite.errorLogs, fmt.Sprintf(format, a...))
		},
	)
	suite.resetErrorLogsOverload = resetLogError
}

func (suite *StreamConnProcessTestSuite) TearDownTest() {
	suite.errorLogs = nil
	suite.resetErrorLogsOverload()
}

func TestStreamConnProcessTestSuite(t *testing.T) {
	suite.Run(t, &StreamConnProcessTestSuite{})
}

func (suite *StreamConnProcessTestSuite) TestNewStreamConnProcess() {
	is := is.New(suite.T())

	testConn := mockCameraConn{}
	readFrames := make(chan video.Frame)
	proc := process.NewStreamConnProcess(&testConn, readFrames)
	is.True(proc != nil)
}

func (suite *StreamConnProcessTestSuite) TestStreamConnProcessReadsFramesFromConn() {
	is := is.New(suite.T())

	clipFrameCount := 36
	testConn := mockCameraConn{isOpen: true, framesToRead: make([]mockFrame, clipFrameCount)}
	readFrames := make(chan video.Frame)
	proc := process.NewStreamConnProcess(&testConn, readFrames)

	proc.Start()
	timeout := time.After(3 * time.Second)
	// readClipsCount := 0
	// readFrameProcLoop:
	readFrameCount := 0
readFrameProcLoop:
	for {
		f := <-readFrames
		select {
		case <-timeout:
			break readFrameProcLoop
		default:
			is.True(f != nil)
			readFrameCount++
			if readFrameCount+1 >= clipFrameCount {
				break readFrameProcLoop
			}
			continue
		}
	}
	proc.Stop()
	proc.Wait()
}
