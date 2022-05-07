package mocks

import (
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
	"github.com/tauraamui/xerror"
)

type Options struct {
	UntrackedFrames bool
	IsOpen          bool
}

func NewCamConn(opts Options) camera.Connection {
	return &mockCameraConn{
		mockOptions: opts,
		isOpen:      opts.IsOpen,
	}
}

type mockFrame struct {
	data          []byte
	width, height int
	isOpen        bool
	isClosing     bool
	onClose       func()
}

func (m *mockFrame) Timestamp() int64 { return 0 }

func (m *mockFrame) DataRef() interface{} {
	return m.data
}

func (m *mockFrame) Dimensions() videoframe.Dimensions {
	return videoframe.Dimensions{W: m.width, H: m.height}
}

func (m *mockFrame) ToBytes() []byte { return nil }

func (m *mockFrame) Close() {
	m.isOpen = false
	m.isClosing = true
	if m.onClose != nil {
		m.onClose()
	}
}

type mockCameraConn struct {
	mockOptions         Options
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
	readFunc            func() (videoframe.Frame, error)
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

func (m *mockCameraConn) Read() (frame videoframe.Frame, err error) {
	if m.onPostRead != nil {
		defer m.onPostRead()
	}

	if m.readFunc != nil {
		return m.readFunc()
	}

	if m.mockOptions.UntrackedFrames {
		return &mockFrame{}, nil
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
