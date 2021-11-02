package repos_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRepos(t *testing.T) {
	t.Skip()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Repos Suite")
}
