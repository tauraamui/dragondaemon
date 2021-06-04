package data_test

import (
	"errors"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	data "github.com/tauraamui/dragondaemon/pkg/database"
)

type testPasswordPromptReader struct {
	testPassword string
	testError    error
}

func (t testPasswordPromptReader) ReadPassword() ([]byte, error) {
	return []byte(t.testPassword), t.testError
}

var _ = Describe("Data", func() {
	Context("Setup run against blank file system", func() {
		It("Should create full file path for DB", func() {
			resetFS := data.OverloadFS(afero.NewMemMapFs())
			defer resetFS()

			resetPromptReader := data.OverloadPromptReader(
				strings.NewReader("testadmin\n"),
			)
			defer resetPromptReader()

			resetPasswordPromptReader := data.OverloadPasswordPromptReader(
				testPasswordPromptReader{
					testPassword: "testpassword",
				},
			)
			defer resetPasswordPromptReader()

			err := data.Setup()
			Expect(err).To(BeNil())
		})
	})

	It("Should return error from setup due to path resolution failure", func() {
		reset := data.OverloadUC(func() (string, error) {
			return "", errors.New("test cache dir error")
		})
		defer reset()

		err := data.Setup()

		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(Equal("unable to resolve dd.db database file location: test cache dir error"))
	})
})
