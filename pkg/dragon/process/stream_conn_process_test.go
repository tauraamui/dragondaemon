package process_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/stretchr/testify/suite"
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

func TestNewStreamConnProcess(t *testing.T) {
	is := is.New(t)

	testConn := mockCameraConn{}
	readFrames := make(chan video.Frame)
	proc := process.NewStreamConnProcess(&testConn, readFrames)
	is.True(proc != nil)
}

func TestStreamConnProcessReadsFramesFromConn(t *testing.T) {
	is := is.New(t)

	clipsCount := 36
	testConn := mockCameraConn{isOpen: true, framesToRead: make([]mockFrame, clipsCount)}
	readFrames := make(chan video.Frame)
	proc := process.NewStreamConnProcess(&testConn, readFrames)

	proc.Start()

	timeout := time.After(3 * time.Second)
readFramesProcLoop:
	for i := 0; i < clipsCount; i++ {
		select {
		case <-timeout:
			t.Fatal("test timed out took longer than 3 seconds to read frames")
			break readFramesProcLoop
		default:
			f := <-readFrames
			is.True(f != nil)
		}
	}

	proc.Stop()
}
