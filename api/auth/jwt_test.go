package auth

import (
	. "github.com/onsi/ginkgo"
	"github.com/tacusci/logging/v2"
)

var _ = Describe("Auth", func() {
	existingLoggingLevel := logging.CurrentLoggingLevel

	BeforeEach(func() {
		logging.CurrentLoggingLevel = logging.SilentLevel
	})

	AfterEach(func() {
		logging.CurrentLoggingLevel = existingLoggingLevel
	})
})
