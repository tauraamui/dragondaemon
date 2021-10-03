package process_test

import (
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/dragon/process"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type mockClipWriter struct {
	writtenClips []video.Clip
	writeErr     error
}

func (m *mockClipWriter) Write(clip video.Clip) error {
	m.writtenClips = append(m.writtenClips, clip)
	return m.writeErr
}

func (m *mockClipWriter) hasWrittenClip(is *is.I, clip video.Clip) bool {
	is.Helper()
	for _, c := range m.writtenClips {
		if c == clip {
			return true
		}
	}
	return false
}

func TestNewPersistClipProcess(t *testing.T) {
	is := is.New(t)

	testWriter := mockClipWriter{}
	clipsToWrite := make(chan video.Clip)
	proc := process.NewPersistClipProcess(clipsToWrite, &testWriter)
	is.True(proc != nil)
}

func TestPersistClipProcessWritesClips(t *testing.T) {
	is := is.New(t)

	clip := video.NewClip("/testroot", 30)
	testWriter := mockClipWriter{}
	clipsToWrite := make(chan video.Clip)
	proc := process.NewPersistClipProcess(clipsToWrite, &testWriter)

	proc.Start()

	clipsToWrite <- clip

	proc.Stop()
	proc.Wait()

	is.True(testWriter.hasWrittenClip(is, clip))
}

func TestPersistClipProcessContinuesIfReaderDelayed(t *testing.T) {
	is := is.New(t)

	clip := video.NewClip("/testroot", 30)
	testWriter := mockClipWriter{}
	clipsToWrite := make(chan video.Clip)
	proc := process.NewPersistClipProcess(clipsToWrite, &testWriter)

	proc.Start()

	time.Sleep(10 * time.Millisecond)

	clipsToWrite <- clip

	proc.Stop()
	proc.Wait()

	is.True(testWriter.hasWrittenClip(is, clip))
}
