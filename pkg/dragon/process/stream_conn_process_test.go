package process_test

import (
	"errors"
	"fmt"
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
	b := broadcast.New(0)

	totalFramesCount := 64
	frames := make([]mockFrame, totalFramesCount)
	testConn := mockCameraConn{
		isOpen: true, framesToRead: frames, schedule: schedule.NewSchedule(schedule.Week{}),
	}

	readFrames := make(chan video.Frame, totalFramesCount)

	timeout := time.After(3 * time.Second)
	startReading := make(chan interface{})
	doneReading := make(chan []error)
	readFrameCount := 0
	loopCount := 0

	go func(b *broadcast.Broadcaster, s chan interface{}, d chan []error, fc chan video.Frame, tc int, rc, lc *int) {
		defer close(d)
		<-s
		errs := []error{}
		reachedTotal := false
	readFrameProcLoop:
		for {
			time.Sleep(1 * time.Microsecond)
			select {
			case <-timeout:
				errs = append(errs, errors.New("test timeout 3s limit exceeded"))
				break readFrameProcLoop
			case f := <-fc:
				f.Close()
				*lc++
				*rc++
				if *rc == tc/2 {
					b.Send(process.CAM_SWITCHED_OFF_EVT)
				}
			default:
				*lc++
				if *lc > tc {
					*lc = tc
				}
				if *lc != 0 && *lc == tc {
					reachedTotal = true
					break readFrameProcLoop
				}
			}
		}
		if !reachedTotal {
			errs = append(errs, fmt.Errorf("counts did not reach total: [%dr/%dt]", *lc, tc))
		}
		d <- errs
	}(b, startReading, doneReading, readFrames, totalFramesCount, &readFrameCount, &loopCount)

	is := is.New(suite.T())

	proc := process.NewStreamConnProcess(b.Listen(), &testConn, readFrames)
	close(startReading)

	proc.Setup().Start()

	errs := <-doneReading
	ec := len(errs)
	is.New(suite.T())

	is.Equal(loopCount/2, readFrameCount)
	for i := 0; i < ec; i++ {
		if err := errs[i]; err != nil {
			fmt.Printf("err: %s\n", err.Error())
			suite.T().Fail()
		}
	}
}

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
