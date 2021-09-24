package process_test

import (
	"errors"

	"github.com/tauraamui/dragondaemon/pkg/video"
)

type mockCameraConn struct {
	uuid                string
	title               string
	persistLocation     string
	fullPersistLocation string
	maxClipAgeDays      int
	fps                 int
	spc                 int
	frameReadIndex      int
	framesToRead        []video.Frame
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

func (m *mockCameraConn) Read() (video.Frame, error) {
	m.frameReadIndex++
	if m.frameReadIndex >= len(m.framesToRead) {
		return nil, errors.New("run out of frames to read")
	}
	return m.framesToRead[m.frameReadIndex], m.readErr
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
