package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tauraamui/dragondaemon/database/models"
)

var _ = Describe("User", func() {
	Context("With new empty instance", func() {
		var user models.User

		BeforeEach(func() {
			user = models.User{}
		})

		It("Should generate new UUID and encrypt authhash plain", func() {
			err := user.BeforeCreate(nil)

			Expect(err).To(BeNil())
			Expect(user.UUID).ToNot(BeEmpty())
		})
	})

	Context("With new user of given name with plain password", func() {
		var user models.User

		BeforeEach(func() {
			user = models.User{
				Name:     "test-user-account",
				AuthHash: "test-user-password",
			}
			err := user.BeforeCreate(nil)
			Expect(err).To(BeNil())
		})

		It("Should generate new UUID and encrypt plain password", func() {
			Expect(user.UUID).ToNot(BeEmpty())
			Expect(user.Name).To(Equal("test-user-account"))
			Expect(user.AuthHash).To(ContainSubstring("$2a$10$"))
		})

		It("Should be able to match plain version to encrypted version of password", func() {
			err := user.ComparePassword("test-user-password")
			Expect(err).To(BeNil())
		})
	})
})
