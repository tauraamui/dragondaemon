package process_test

import (
	"fmt"
	"testing"

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

	testConn := mockCameraConn{
		framesToRead: make([]mockFrame, 10),
	}

	for i := 0; i < 10; i++ {
		frame, err := testConn.Read()
		is.NoErr(err)
		is.True(frame != nil)
	}
}
