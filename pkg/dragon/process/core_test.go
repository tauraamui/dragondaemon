package process

import (
	"errors"
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/video"
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

func (m *mockFrame) Dimensions() video.FrameDimension {
	return video.FrameDimension{W: m.width, H: m.height}
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

func (m *mockCameraConn) SPC() int {
	return m.spc
}

func (m *mockCameraConn) Read() (frame video.Frame, err error) {
	if m.onPostRead != nil {
		defer m.onPostRead()
	}
	if m.frameReadIndex+1 >= len(m.framesToRead) {
		return nil, errors.New("run out of frames to read")
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

func (m *mockCameraConn) reset() {
	m.frameReadIndex = 0
}

type mockClipWriter struct {
	writtenClips []video.Clip
	writeErr     error
}

func (m *mockClipWriter) Write(clip video.Clip) error {
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
	onStart func()
	onStop  func()
	onWait  func()
}

func (m *mockProc) Setup() {}

func (m *mockProc) Start() {
	if m.onStart != nil {
		m.onStart()
	}
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

	streamProcCalled := false
	onStreamProcStart := func() { streamProcCalled = true }
	generateProcCalled := false
	onGenerateProcStart := func() { generateProcCalled = true }
	persistProcCalled := false
	onPersistProcStart := func() { persistProcCalled = true }

	proc.streamProcess = &mockProc{onStart: onStreamProcStart}
	proc.generateClips = &mockProc{onStart: onGenerateProcStart}
	proc.persistClips = &mockProc{onStart: onPersistProcStart}

	proc.Start()

	is.True(streamProcCalled)
	is.True(generateProcCalled)
	is.True(persistProcCalled)
}

func TestCoreProcessStop(t *testing.T) {
	is := is.New(t)
	conn := mockCameraConn{}
	writer := mockClipWriter{}
	proc := NewCoreProcess(&conn, &writer).(*persistCameraToDisk)

	streamProcCalled := false
	onStreamProcStop := func() { streamProcCalled = true }
	generateProcCalled := false
	onGenerateProcStop := func() { generateProcCalled = true }
	persistProcCalled := false
	onPersistProcStop := func() { persistProcCalled = true }

	proc.streamProcess = &mockProc{onStop: onStreamProcStop}
	proc.generateClips = &mockProc{onStop: onGenerateProcStop}
	proc.persistClips = &mockProc{onStop: onPersistProcStop}

	proc.Stop()

	is.True(streamProcCalled)
	is.True(generateProcCalled)
	is.True(persistProcCalled)
}

func TestCoreProcessWait(t *testing.T) {
	is := is.New(t)
	conn := mockCameraConn{}
	writer := mockClipWriter{}
	proc := NewCoreProcess(&conn, &writer).(*persistCameraToDisk)

	streamProcCalled := false
	onStreamProcWait := func() { streamProcCalled = true }
	generateProcCalled := false
	onGenerateProcWait := func() { generateProcCalled = true }
	persistProcCalled := false
	onPersistProcWait := func() { persistProcCalled = true }

	proc.streamProcess = &mockProc{onWait: onStreamProcWait}
	proc.generateClips = &mockProc{onWait: onGenerateProcWait}
	proc.persistClips = &mockProc{onWait: onPersistProcWait}

	proc.Wait()

	is.True(streamProcCalled)
	is.True(generateProcCalled)
	is.True(persistProcCalled)
}
