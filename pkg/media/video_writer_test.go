package media_test

import (
	"os"

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

	Context("", func() {
		It("Should return a new ", func() {
			Expect(mockFs.MkdirAll("/testroot/clips/TestConnection", os.ModeDir|os.ModePerm)).To(BeNil())
		})
	})
})
