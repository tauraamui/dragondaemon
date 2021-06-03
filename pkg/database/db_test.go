package data_test

import (
	"errors"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	data "github.com/tauraamui/dragondaemon/pkg/database"
)

var _ = Describe("Data", func() {
	Context("Setup run against blank file system", func() {
		It("Should create full file path for DB", func() {
			resetFS := data.OverloadFS(afero.NewMemMapFs())
			defer resetFS()

			resetPromptReader := data.OverloadPromptReader(
				strings.NewReader("testadmin\ntestpassword\n"),
			)
			defer resetPromptReader()

			err := data.Setup()
			Expect(err).ToNot(BeNil())
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
