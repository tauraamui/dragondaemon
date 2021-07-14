package media_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/tacusci/logging/v2"

	"github.com/tauraamui/dragondaemon/pkg/media"
)

const testRootClipsPath = "/testroot/clips/TestVideoWriter"

var _ = Describe("VideoWriter", func() {

	existingLoggingLevel := logging.CurrentLoggingLevel

	var mockFs afero.Fs = nil
	var resetFs func() = nil

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

	Context("Mock video writer", func() {

		It("Checks mock video writer returns expected results", func() {
			mockVidWriter, err := media.OpenVideoWriter(
				filepath.Join(testRootClipsPath, "testclip.mp4"), "avc1.4d001e", 30, 10, 10, true,
			)
			Expect(mockVidWriter).ToNot(BeNil())
			Expect(err).To(BeNil())
			Expect(mockVidWriter.IsOpened()).To(BeTrue())
			Expect(mockVidWriter.Close()).To(BeNil())
		})
	})
})
