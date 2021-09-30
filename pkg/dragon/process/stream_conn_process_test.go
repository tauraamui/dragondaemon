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
	frames := []mockFrame{}
	for i := 0; i < clipFrameCount; i++ {
		frames = append(frames, mockFrame{
			data: []byte{0x0A << i},
		})
	}
	testConn := mockCameraConn{isOpen: true, framesToRead: frames}
	// make test channel buffered to allow the send
	// routine to optionally send, and our test reciever
	// to optionally recieve without blocking so the loop
	// proceeds and the timeout is checked
	readFrames := make(chan video.Frame, 3)
	proc := process.NewStreamConnProcess(&testConn, readFrames)

	proc.Start()
	timeout := time.After(3 * time.Second)
	readFrameCount := 0
readFrameProcLoop:
	for {
		select {
		case <-timeout:
			break readFrameProcLoop
		case f := <-readFrames:
			is.True(f != nil)
			data, ok := f.DataRef().([]byte)
			is.True(ok)
			is.Equal([]byte{0x0A << readFrameCount}, data)
			readFrameCount++
			if readFrameCount+1 >= clipFrameCount {
				break readFrameProcLoop
			}
		}
	}
	proc.Stop()
	proc.Wait()
}
