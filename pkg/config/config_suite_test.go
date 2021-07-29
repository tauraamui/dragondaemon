package config_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tacusci/logging/v2"
)

func TestConfig(t *testing.T) {
	existingLoggingLevel := logging.CurrentLoggingLevel
	logging.CurrentLoggingLevel = logging.SilentLevel
	// make this as defer, in case a panic is handled by Ginkgo
	defer func() { logging.CurrentLoggingLevel = existingLoggingLevel }()

	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}
