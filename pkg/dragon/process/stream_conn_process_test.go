package process_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/stretchr/testify/suite"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/broadcast"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"github.com/tauraamui/dragondaemon/pkg/dragon/process"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
	"github.com/tauraamui/dragondaemon/pkg/xis"
	"github.com/tauraamui/xerror"
)

func overloadErrorLog(overload func(string, ...interface{})) func() {
	logErrorRef := log.Error
	log.Error = overload
	return func() { log.Error = logErrorRef }
}

type StreamConnProcessTestSuite struct {
	suite.Suite
	timeSecondOffset       int
	resetErrorLogsOverload func()
	errorLogs              []string
	onPostErrorLog         func()
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
			if suite.onPostErrorLog != nil {
				suite.onPostErrorLog()
			}
		},
	)
	suite.resetErrorLogsOverload = resetLogError

	suite.timeSecondOffset = 0
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

	testConn := mockCameraConn{schedule: schedule.NewSchedule(schedule.Week{})}
	readFrames := make(chan video.Frame)
	proc := process.NewStreamConnProcess(broadcast.New(0).Listen(), &testConn, readFrames)
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
	testConn := mockCameraConn{
		isOpen: true, framesToRead: frames, schedule: schedule.NewSchedule(schedule.Week{}),
	}
	// make test channel buffered to allow the send
	// routine to optionally send, and our test reciever
	// to optionally recieve without blocking so the loop
	// proceeds and the timeout is checked
	readFrames := make(chan video.Frame, 3)
	proc := process.NewStreamConnProcess(broadcast.New(0).Listen(), &testConn, readFrames)

	proc.Setup().Start()
	timeout := time.After(3 * time.Second)
	readFrameCount := 0
readFrameProcLoop:
	for {
		select {
		case <-timeout:
			suite.T().Fatal("test timeout 3s limit exceeded")
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

func (suite *StreamConnProcessTestSuite) TestStreamConnProcessStopsReadingFramesAfterCamOffEvent() {
	maxLoopCount := 64
	oc := mutexCounter{}
	isOpen := func() (open bool) {
		open = true
		// DEVNOTE(tauraamui):
		// yeah yeah this should never happen, don't
		// judge me I just want to wish upon a star
		// and banish all multithreaded based issues
		// with non-determinism in these tests...
		if oc.v() > maxLoopCount {
			oc.set(maxLoopCount)
			return
		}
		if oc.v() == maxLoopCount {
			return
		}
		oc.incr()
		return
	}

	rc := mutexCounter{}
	connRead := func() (video.Frame, error) {
		rc.incr()
		return &mockFrame{}, nil
	}
	testConn := mockCameraConn{readFunc: connRead, isOpenFunc: isOpen}
	fc := make(chan video.Frame)

	b := broadcast.New(0)
	proc := process.NewStreamConnProcess(b.Listen(), &testConn, fc)

	is := is.New(suite.T())
	<-proc.Setup().Start()

	err := callW3sTimeout(func() {
		for {
			time.Sleep(1 * time.Microsecond)
			if rc.v() >= maxLoopCount/2 {
				b.Send(process.CAM_SWITCHED_OFF_EVT)
			}
			if oc.v() >= maxLoopCount {
				break
			}
		}
	})
	is.NoErr(err)

	is.Equal(oc.v(), maxLoopCount)
	is.True(rc.v() == maxLoopCount/2 || rc.v() == (maxLoopCount/2)+1) // read frame count should be half of total loop count

	err = callW3sTimeout(func() { proc.Stop(); proc.Wait() })
	is.NoErr(err)
}

type mutexCounter struct {
	mu sync.Mutex
	c  int
}

func (c *mutexCounter) set(v int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.c = v
}

func (c *mutexCounter) v() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c
}

func (c *mutexCounter) incr() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.c++
}

func callW3sTimeout(f func()) error {
	return callWTimeout(f, time.After(3*time.Second), "test timeout 3s limit exceeded")
}

func callWTimeout(f func(), t <-chan time.Time, errmsg string) error {
	done := make(chan interface{})
	go func(d chan interface{}, f func()) {
		defer close(d)
		f()
	}(done, f)

	for {
		select {
		case <-t:
			return xerror.New(errmsg)
		case <-done:
			return nil
		}
	}
}

func (suite *StreamConnProcessTestSuite) TestStreamConnProcessUnableToReturnFrameDueToNoReader() {
	is := is.New(suite.T())

	closedFramesCount := 0
	incrCloseCount := func() { closedFramesCount++ }
	firstFrame := mockFrame{
		onClose: incrCloseCount,
	}
	secondFrame := mockFrame{
		onClose: incrCloseCount,
	}
	thirdFrame := mockFrame{
		onClose: incrCloseCount,
	}
	forthFrame := mockFrame{
		onClose: incrCloseCount,
	}
	fithFrame := mockFrame{
		onClose: incrCloseCount,
	}
	sixthFrame := mockFrame{
		onClose: incrCloseCount,
	}

	readFrameCount := 0
	testConn := mockCameraConn{
		isOpen: true, onPostRead: func() { readFrameCount++ },
		framesToRead: []mockFrame{
			firstFrame, secondFrame, thirdFrame, forthFrame, fithFrame, sixthFrame,
		},
		schedule: schedule.NewSchedule(schedule.Week{}),
	}

	readFrames := make(chan video.Frame, 2)
	proc := process.NewStreamConnProcess(broadcast.New(0).Listen(), &testConn, readFrames)

	proc.Setup().Start()
	timeout := time.After(3 * time.Second)
checkFrameReadCountLoop:
	for {
		select {
		case <-timeout:
			suite.T().Fatal("test timeout 3s limit exceeded")
			break checkFrameReadCountLoop
		default:
			if readFrameCount >= 6 {
				break checkFrameReadCountLoop
			}
		}
	}
	proc.Stop()
	proc.Wait()

	is.Equal(closedFramesCount, 3)
}

func (suite *StreamConnProcessTestSuite) TestStreamConnProcessUnableToReadError() {
	is := is.New(suite.T())
	testConn := mockCameraConn{
		isOpen:   true,
		readErr:  xerror.New("testing unable to read from mock camera stream"),
		schedule: schedule.NewSchedule(schedule.Week{}),
	}

	readFrames := make(chan video.Frame)
	proc := process.NewStreamConnProcess(broadcast.New(0).Listen(), &testConn, readFrames)

	suite.onPostErrorLog = func() {
		proc.Stop()
	}

	proc.Setup().Start()
	proc.Wait()

	is.True(xis.Contains(
		suite.errorLogs,
		"Unable to retrieve frame: run out of frames to read. Auto re-connecting is not yet implemented",
	))
}
