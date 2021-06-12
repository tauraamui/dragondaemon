package media_test

import (
	"io"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/config"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"github.com/tauraamui/dragondaemon/pkg/media"
	"gocv.io/x/gocv"
)

type testMockVideoCapture struct {
}

// SetP doesn't do anything it exists to satisfy VideoCapturable interface
func (tmvc *testMockVideoCapture) SetP(_ *gocv.VideoCapture) {}

// IsOpened always returns true.
func (tmvc *testMockVideoCapture) IsOpened() bool {
	return true
}

func (tmvc *testMockVideoCapture) Read(m *gocv.Mat) bool {
	return false
}

func (tmvc *testMockVideoCapture) Close() error {
	return nil
}

var _ = Describe("Connection", func() {
	existingLoggingLevel := logging.CurrentLoggingLevel

	Context("NewConnection", func() {
		var resetFs func() = nil
		var mockFs afero.Fs = nil

		BeforeEach(func() {
			logging.CurrentLoggingLevel = logging.SilentLevel
			mockFs = afero.NewMemMapFs()
			resetFs = media.OverloadFS(mockFs)
		})

		AfterEach(func() {
			logging.CurrentLoggingLevel = existingLoggingLevel
			resetFs()
			mockFs = nil
		})

		It("Should return a new connection instance", func() {
			mockFs.MkdirAll("/testroot/clips/TestConnection", os.ModeDir|os.ModePerm)
			conn := media.NewConnection(
				"TestConnection",
				media.ConnectonSettings{
					PersistLocation: "/testroot/clips",
					FPS:             30,
					SecondsPerClip:  2,
					Schedule:        schedule.Schedule{},
					Reolink:         config.ReolinkAdvanced{Enabled: false},
				},
				&testMockVideoCapture{},
				"fake-stream-addr",
			)

			Expect(conn).ToNot(BeNil())
			Expect(conn.UUID()).ToNot(BeEmpty())
			Expect(conn.Title()).To(Equal("TestConnection"))
			_, _, err := conn.SizeOnDisk()
			Expect(err).To(MatchError(io.EOF))
		})
	})
})
