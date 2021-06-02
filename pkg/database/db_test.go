package data_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	data "github.com/tauraamui/dragondaemon/pkg/database"
)

var _ = Describe("Data", func() {
	It("Should return error from setup due to path resolution failure", func() {
		data.OverloadUC(func() (string, error) {
			return "", errors.New("test cache dir error")
		})
		err := data.Setup()

		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(Equal("unable to resolve dd.db database file location: test cache dir error"))
	})
})
