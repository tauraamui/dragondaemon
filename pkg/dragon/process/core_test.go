package process

import (
	"sync"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/broadcast"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"github.com/tauraamui/dragondaemon/pkg/video/videoclip"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
	"github.com/tauraamui/xerror"
)

type mockFrame struct {
	data          []byte
	width, height int
	isOpen        bool
	isClosing     bool
	onClose       func()
}

func (m *mockFrame) DataRef() interface{} {
	return m.data
}

func (m *mockFrame) Dimensions() videoframe.Dimensions {
	return videoframe.Dimensions{W: m.width, H: m.height}
}

func (m *mockFrame) Close() {
	m.isOpen = false
	m.isClosing = true
	if m.onClose != nil {
		m.onClose()
	}
}

type mockCameraConn struct {
	uuid                string
	title               string
	persistLocation     string
	fullPersistLocation string
	maxClipAgeDays      int
	fps                 int
	schedule            schedule.Schedule
	spc                 int
	frameReadIndex      int
	framesToRead        []mockFrame
	onPostRead          func()
	readErr             error
	isOpen              bool
	isClosing           bool
	closeErr            error
}

func (m *mockCameraConn) UUID() string {
	return m.uuid
}

func (m *mockCameraConn) Title() string {
	return m.title
}

func (m *mockCameraConn) PersistLocation() string {
	return m.persistLocation
}

func (m *mockCameraConn) FullPersistLocation() string {
	return m.fullPersistLocation
}

func (m *mockCameraConn) MaxClipAgeDays() int {
	return m.maxClipAgeDays
}

func (m *mockCameraConn) FPS() int {
	return m.fps
}

func (m *mockCameraConn) Schedule() schedule.Schedule {
	if m.schedule == nil {
		m.schedule = schedule.NewSchedule(schedule.Week{})
	}
	return m.schedule
}

func (m *mockCameraConn) SPC() int {
	return m.spc
}

func (m *mockCameraConn) Read() (frame videoframe.Frame, err error) {
	if m.onPostRead != nil {
		defer m.onPostRead()
	}
	if m.frameReadIndex+1 >= len(m.framesToRead) {
		return nil, xerror.New("run out of frames to read")
	}
	frame, err = &m.framesToRead[m.frameReadIndex], m.readErr
	m.frameReadIndex++
	return
}

func (m *mockCameraConn) IsOpen() bool {
	return m.isOpen
}

func (m *mockCameraConn) IsClosing() bool {
	return m.isClosing
}

func (m *mockCameraConn) Close() error {
	return m.closeErr
}

type mockClipWriter struct {
	writtenClips []videoclip.NoCloser
	writeErr     error
}

func (m *mockClipWriter) Write(clip videoclip.NoCloser) error {
	m.writtenClips = append(m.writtenClips, clip)
	return m.writeErr
}

func TestNewCoreProcess(t *testing.T) {
	is := is.New(t)
	conn := mockCameraConn{}
	writer := mockClipWriter{}
	proc := NewCoreProcess(&conn, &writer)

	is.True(proc != nil)
}

func TestCoreProcessSetup(t *testing.T) {
	is := is.New(t)
	conn := mockCameraConn{}
	writer := mockClipWriter{}
	proc := NewCoreProcess(&conn, &writer).(*persistCameraToDisk)

	proc.Setup()
	is.True(proc.streamProcess != nil)
	is.True(proc.generateClips != nil)
	is.True(proc.persistClips != nil)
}

type mockProc struct {
	started chan struct{}
	onStart func()
	onStop  func()
	onWait  func()
}

func (m *mockProc) Setup() Process {
	return m
}

func (m *mockProc) Start() <-chan struct{} {
	if m.onStart != nil {
		m.onStart()
	}
	if m.started == nil {
		m.started = make(chan struct{})
	}
	defer close(m.started)
	return m.started
}

func (m *mockProc) Stop() {
	if m.onStop != nil {
		m.onStop()
	}
}

func (m *mockProc) Wait() {
	if m.onWait != nil {
		m.onWait()
	}
}

func TestCoreProcessStart(t *testing.T) {
	is := is.New(t)
	conn := mockCameraConn{}
	writer := mockClipWriter{}
	proc := NewCoreProcess(&conn, &writer).(*persistCameraToDisk)

	monitorCamStateProcCalled := false
	onMonitorCamStateProcStart := func() { monitorCamStateProcCalled = true }
	streamProcCalled := false
	onStreamProcStart := func() { streamProcCalled = true }
	generateProcCalled := false
	onGenerateProcStart := func() { generateProcCalled = true }
	persistProcCalled := false
	onPersistProcStart := func() { persistProcCalled = true }

	proc.monitorCameraOnState = &mockProc{onStart: onMonitorCamStateProcStart}
	proc.streamProcess = &mockProc{onStart: onStreamProcStart}
	proc.generateClips = &mockProc{onStart: onGenerateProcStart}
	proc.persistClips = &mockProc{onStart: onPersistProcStart}

	proc.Start()

	is.True(monitorCamStateProcCalled)
	is.True(streamProcCalled)
	is.True(generateProcCalled)
	is.True(persistProcCalled)
}

func TestCoreProcessStop(t *testing.T) {
	is := is.New(t)
	conn := mockCameraConn{}
	writer := mockClipWriter{}
	proc := NewCoreProcess(&conn, &writer).(*persistCameraToDisk)

	monitorCamStateProcCalled := false
	onMonitorCamStateProcStop := func() { monitorCamStateProcCalled = true }
	streamProcCalled := false
	onStreamProcStop := func() { streamProcCalled = true }
	generateProcCalled := false
	onGenerateProcStop := func() { generateProcCalled = true }
	persistProcCalled := false
	onPersistProcStop := func() { persistProcCalled = true }

	proc.monitorCameraOnState = &mockProc{onStop: onMonitorCamStateProcStop}
	proc.streamProcess = &mockProc{onStop: onStreamProcStop}
	proc.generateClips = &mockProc{onStop: onGenerateProcStop}
	proc.persistClips = &mockProc{onStop: onPersistProcStop}

	proc.Stop()

	is.True(monitorCamStateProcCalled)
	is.True(streamProcCalled)
	is.True(generateProcCalled)
	is.True(persistProcCalled)
}

func TestCoreProcessWait(t *testing.T) {
	is := is.New(t)
	conn := mockCameraConn{}
	writer := mockClipWriter{}
	proc := NewCoreProcess(&conn, &writer).(*persistCameraToDisk)

	monitorCamStateProcCalled := false
	onMonitorCamStateProcWait := func() { monitorCamStateProcCalled = true }
	streamProcCalled := false
	onStreamProcWait := func() { streamProcCalled = true }
	generateProcCalled := false
	onGenerateProcWait := func() { generateProcCalled = true }
	persistProcCalled := false
	onPersistProcWait := func() { persistProcCalled = true }

	proc.monitorCameraOnState = &mockProc{onWait: onMonitorCamStateProcWait}
	proc.streamProcess = &mockProc{onWait: onStreamProcWait}
	proc.generateClips = &mockProc{onWait: onGenerateProcWait}
	proc.persistClips = &mockProc{onWait: onPersistProcWait}

	proc.Wait()

	is.True(monitorCamStateProcCalled)
	is.True(streamProcCalled)
	is.True(generateProcCalled)
	is.True(persistProcCalled)
}

func TestSendEventOnCameraStateChange(t *testing.T) {
	tm := timeMachine{
		offset:   new(int),
		baseTime: time.Date(2021, 3, 17, 13, 0, 0, 0, time.UTC),
	}

	r := overloadTimeNow(tm.timeNowQuery)
	defer r()
	schedule.TODAY = time.Date(2021, 3, 17, 0, 0, 0, 0, time.UTC)

	offTimeSeconds := 18
	b := broadcast.New(0)
	conn := mockCameraConn{
		isOpen: true,
		schedule: schedule.NewSchedule(schedule.Week{
			Wednesday: schedule.OnOffTimes{
				Off: testTimePtr(args{
					hour: 13, minute: 0, second: offTimeSeconds,
				}),
			},
		}),
	}

	proc := New(Settings{
		WaitForShutdownMsg: "",
		Process:            sendEvtOnCameraStateChange(b, &conn, time.Millisecond),
	})

	proc.Setup().Start()
	l := b.Listen()
	err := callW3sTimeout(func() error {
		for msg := range l.Ch {
			if evt, ok := msg.(Event); ok && evt == CAM_SWITCHED_OFF_EVT {
				v := tm.v()
				if v == offTimeSeconds+1 {
					return nil
				}
				return xerror.Errorf("seconds %d do not match target %d", v, offTimeSeconds)
			}
		}
		return xerror.New("early error return did not occur")
	})

	is := is.New(t)
	is.NoErr(err)

	proc.Stop()
	proc.Wait()
}

func overloadTimeNow(o func() time.Time) func() {
	ref := TimeNow
	TimeNow = o
	return func() { TimeNow = ref }
}

type timeMachine struct {
	mu       sync.Mutex
	offset   *int
	baseTime time.Time
}

func (t *timeMachine) timeNowQuery() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	*t.offset++
	return t.baseTime.Add(time.Second * time.Duration(*t.offset))
}

func (t *timeMachine) v() int {
	return *t.offset
}

func callW3sTimeout(f func() error) error {
	return callWTimeout(f, time.After(3*time.Second), "test timeout 3s limit exceeded")
}

func callWTimeout(f func() error, t <-chan time.Time, errmsg string) error {
	done := make(chan struct{})
	err := make(chan error)
	go func(d chan struct{}, f func() error, e chan error) {
		defer close(d)
		defer close(err)
		err <- f()
	}(done, f, err)

	for {
		select {
		case <-t:
			return xerror.New(errmsg)
		case e := <-err:
			return e
		case <-done:
			return nil
		}
	}
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
