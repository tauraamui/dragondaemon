package process_test

import (
	"errors"
	"fmt"
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
	proc := process.NewStreamConnProcess(broadcast.New(0), &testConn, readFrames)
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
	proc := process.NewStreamConnProcess(broadcast.New(0), &testConn, readFrames)

	proc.Start()
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

func (suite *StreamConnProcessTestSuite) TestStreamConnProcessStopsReadingFramesAfterCamTurnsOff() {
	process.TimeNow = suite.timeNowQuery
	schedule.TODAY = time.Date(2021, 3, 17, 0, 0, 0, 0, time.UTC)
	suite.baseTime = time.Date(2021, 3, 17, 13, 0, 0, 0, time.UTC)

	is := is.New(suite.T())

	broadcst := broadcast.New(0)
	listener := broadcst.Listen()

	clipFrameCount := 36
	frames := make([]mockFrame, clipFrameCount)
	testConn := mockCameraConn{
		isOpen: true, framesToRead: frames, schedule: schedule.NewSchedule(schedule.Week{
			Wednesday: schedule.OnOffTimes{
				Off: testTimePtr(args{
					hour: 13, minute: 0, second: clipFrameCount / 2,
				}),
			},
		}),
	}

	// make test channel buffered to allow the send
	// routine to optionally send, and our test reciever
	// to optionally recieve without blocking so the loop
	// proceeds and the timeout is checked
	readFrames := make(chan video.Frame, 3)
	proc := process.NewStreamConnProcess(broadcst, &testConn, readFrames)

	proc.Start()
	timeout := time.After(3 * time.Second)
	readFrameCount := 0
	lastIterationGotMsg := false
readFrameProcLoop:
	for {
		select {
		case <-timeout:
			suite.T().Fatal("test timeout 3s limit exceeded")
			break readFrameProcLoop
		case msg := <-listener.Ch:
			if evt, ok := msg.(process.Event); ok && evt == process.CAM_SWITCHED_OFF_EVT {
				lastIterationGotMsg = true
				continue
			}
		case f := <-readFrames:
			is.True(f != nil)
			readFrameCount++
		default:
			if readFrameCount+1 > clipFrameCount/2 {
				if !lastIterationGotMsg {
					suite.T().Fatal("camera off event not ever sent")
				}
				break readFrameProcLoop
			}
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
	proc := process.NewStreamConnProcess(broadcast.New(0), &testConn, readFrames)
	proc.Start()

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
		readErr:  errors.New("testing unable to read from mock camera stream"),
		schedule: schedule.NewSchedule(schedule.Week{}),
	}

	readFrames := make(chan video.Frame)
	proc := process.NewStreamConnProcess(broadcast.New(0), &testConn, readFrames)

	suite.onPostErrorLog = func() {
		proc.Stop()
	}

	proc.Start()
	proc.Wait()

	is.Equal(suite.errorLogs, []string{
		"Unable to retrieve frame: run out of frames to read. Auto re-connecting is not yet implemented",
	})
}

func (suite *StreamConnProcessTestSuite) timeNowQuery() time.Time {
	suite.timeSecondOffset++
	return suite.baseTime.Add(time.Second * time.Duration(suite.timeSecondOffset))
}

// used in place of unavailable named params langauge feature
type args struct {
	date                 timeDate
	hour, minute, second int
}

type timeDate struct {
	year, month, day int
}

func (td timeDate) empty() bool {
	if td.year == 0 || td.month == 0 || td.day == 0 {
		return true
	}
	return false
}

var defaultDate timeDate = timeDate{2021, 9, 13}

func timeFromHoursAndMinutes(td timeDate, hour, minute, second int) schedule.Time {
	return schedule.Time(time.Date(td.year, time.Month(td.month), td.day, hour, minute, second, 0, time.UTC))
}

func testTime(a args) schedule.Time {
	date := func() timeDate {
		if a.date.empty() {
			return defaultDate
		}
		return a.date
	}()
	return timeFromHoursAndMinutes(date, a.hour, a.minute, a.second)
}

func testTimePtr(a args) *schedule.Time {
	tt := testTime(a)
	return &tt
}
