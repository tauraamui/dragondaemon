package media_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/tacusci/logging/v2"

	"github.com/tauraamui/dragondaemon/pkg/media"
)

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

	It("Should return a video writer", func() {
		videoWriter, err := media.OpenVideoWriter(
			"testclip.mp4", "avc1.4d001e", 30, 10, 10,
		)

		Expect(videoWriter).ToNot(BeNil())
		Expect(err).To(BeNil())
	})
})
