package media_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMedia(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Media Suite")
}
