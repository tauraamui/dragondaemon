package process_test

import (
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
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
	schedule       schedule.Schedule
	frameReadIndex int
	framesToRead   []mockFrame
	readFunc       func() (videoframe.Frame, error)
	onPostRead     func()
	readErr        error
	isOpenFunc     func() bool
	isOpen         bool
}

func (m *mockCameraConn) Read() (frame videoframe.Frame, err error) {
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
