package repos_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/database/models"
	"github.com/tauraamui/dragondaemon/pkg/database/repos"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _ = Describe("UserRepo", func() {
	existingLoggingLevel := logging.CurrentLoggingLevel
	var (
		mockDBConn *gorm.DB
	)

	BeforeEach(func() {
		logging.CurrentLoggingLevel = logging.SilentLevel

		// handy "hack" to create temp testible DB, open empty SQLite DB in memory
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		Expect(err).To(BeNil())
		Expect(db).ToNot(BeNil())
		mockDBConn = db

		models.AutoMigrate(mockDBConn)
	})

	AfterEach(func() {
		logging.CurrentLoggingLevel = existingLoggingLevel

		db, err := mockDBConn.DB()
		Expect(err).To(BeNil())
		err = db.Close()
		Expect(err).To(BeNil())
	})

	Describe("UserRepository", func() {
		var (
			repo repos.UserRepository
		)

		BeforeEach(func() {
			repo = repos.UserRepository{
				DB: mockDBConn,
			}
		})

		Context("With existing user", func() {
			var existingUserUUID string
			BeforeEach(func() {
				u := models.User{
					Name:     "test-user-account",
					AuthHash: "test-user-password",
				}
				err := repo.Create(&u)
				Expect(err).To(BeNil())
				existingUserUUID = u.UUID
			})

			It("Should find the existing user by it's UUID", func() {
				u, err := repo.FindByUUID(existingUserUUID)
				Expect(err).To(BeNil())
				Expect(u.UUID).To(Equal(existingUserUUID))
				Expect(u.Name).To(Equal("test-user-account"))
			})

			It("Should find the existing user by it's name", func() {
				u, err := repo.FindByName("test-user-account")
				Expect(err).To(BeNil())
				Expect(u.UUID).To(Equal(existingUserUUID))
				Expect(u.Name).To(Equal("test-user-account"))
			})

			It("Should have created the user with an encrypted version of password", func() {
				u, err := repo.FindByUUID(existingUserUUID)
				Expect(err).To(BeNil())
				Expect(u.AuthHash).ToNot(BeEmpty())
				Expect(u.AuthHash).ToNot(Equal("test-user-password"))
			})
		})
	})
})
