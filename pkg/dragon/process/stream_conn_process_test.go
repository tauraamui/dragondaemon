package process_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/broadcast"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
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
	baseTime               time.Time
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
	isOpen := func() bool {
		oc.incr()
		v := oc.v()
		if v > maxLoopCount {
			oc.set(maxLoopCount)
		}
		return true
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
	proc.Setup().Start()

	err := callW3sTimeout(func() {
		for {
			time.Sleep(1 * time.Microsecond)
			if rc.v() == maxLoopCount/2 {
				b.Send(process.CAM_SWITCHED_OFF_EVT)
			}
			if oc.v() == maxLoopCount {
				break
			}
		}
	})
	is.NoErr(err)

	is.Equal(oc.v(), maxLoopCount)
	is.Equal(rc.v(), maxLoopCount/2) // read frame count should be half of total loop count

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
			return errors.New(errmsg)
		case <-done:
			return nil
		}
	}
}

// TODO(tauraamui): actually move this test to the schedule package tests
func (suite *StreamConnProcessTestSuite) TestStreamConnProcessStopsReadingFramesAfterCamTurnsOff() {
	suite.T().Skip()
	process.TimeNow = suite.timeNowQuery
	schedule.TODAY = time.Date(2021, 3, 17, 0, 0, 0, 0, time.UTC)
	suite.baseTime = time.Date(2021, 3, 17, 13, 0, 0, 0, time.UTC)

	broadcst := broadcast.New(0)
	listener := broadcst.Listen()

	clipFrameCount := 36
	frames := make([]mockFrame, clipFrameCount)
	testConn := mockCameraConn{
		isOpen: true, framesToRead: frames, schedule: schedule.NewSchedule(schedule.Week{
			Wednesday: schedule.OnOffTimes{
				// Off: testTimePtr(args{
				// 	hour: 13, minute: 0, second: clipFrameCount / 2,
				// }),
			},
		}),
	}

	// make test channel buffered to allow the send
	// routine to optionally send, and our test reciever
	// to optionally recieve without blocking so the loop
	// proceeds and the timeout is checked
	readFrames := make(chan video.Frame, 3)
	proc := process.NewStreamConnProcess(listener, &testConn, readFrames)

	proc.Setup().Start()
	timeout := time.After(3 * time.Second)
readFrameProcLoop:
	for {
		select {
		case <-timeout:
			suite.T().Fatal("test timeout 3s limit exceeded")
			break readFrameProcLoop
		case msg := <-listener.Ch:
			if evt, ok := msg.(process.Event); ok && evt == process.CAM_SWITCHED_OFF_EVT {
				target := (clipFrameCount / 2) + 1
				if suite.timeSecondOffset != target {
					suite.T().Fatal(
						fmt.Sprintf("camera off sent at wrong time: %ds/%d in", suite.timeSecondOffset, target),
					)
				}
				break readFrameProcLoop
			}
		case <-readFrames:
		}
	}
	proc.Stop()
	proc.Wait()
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
	testConn := mockCameraConn{
		isOpen:   true,
		readErr:  errors.New("testing unable to read from mock camera stream"),
		schedule: schedule.NewSchedule(schedule.Week{}),
	}

	readFrames := make(chan video.Frame)
	proc := process.NewStreamConnProcess(broadcast.New(0).Listen(), &testConn, readFrames)

	suite.onPostErrorLog = func() {
		proc.Stop()
	}

	proc.Setup().Start()
	proc.Wait()

	assert.Contains(
		suite.T(),
		suite.errorLogs,
		"Unable to retrieve frame: run out of frames to read. Auto re-connecting is not yet implemented",
	)
}

func (suite *StreamConnProcessTestSuite) timeNowQuery() time.Time {
	suite.timeSecondOffset++
	return suite.baseTime.Add(time.Second * time.Duration(suite.timeSecondOffset))
}
