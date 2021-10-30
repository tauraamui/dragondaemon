package process_test

import (
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"github.com/tauraamui/dragondaemon/pkg/video"
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
	schedule            schedule.Schedule
	spc                 int
	frameReadIndex      int
	framesToRead        []mockFrame
	readFunc            func() (video.Frame, error)
	onPostRead          func()
	readErr             error
	isOpenFunc          func() bool
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
	return m.schedule
}

func (m *mockCameraConn) SPC() int {
	return m.spc
}

func (m *mockCameraConn) Read() (frame video.Frame, err error) {
	if m.onPostRead != nil {
		defer m.onPostRead()
	}

	if m.readFunc != nil {
		return m.readFunc()
	}

	if m.frameReadIndex+1 >= len(m.framesToRead) {
		return nil, xerror.New("run out of frames to read")
	}
	frame, err = &m.framesToRead[m.frameReadIndex], m.readErr
	m.frameReadIndex++
	return
}

func (m *mockCameraConn) IsOpen() bool {
	if m.isOpenFunc != nil {
		return m.isOpenFunc()
	}
	return m.isOpen
}

func (m *mockCameraConn) IsClosing() bool {
	return m.isClosing
}

func (m *mockCameraConn) Close() error {
	return m.closeErr
}
