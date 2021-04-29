package repos_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tauraamui/dragondaemon/database/models"
	"github.com/tauraamui/dragondaemon/database/repos"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _ = Describe("UserRepo", func() {
	var (
		mockDBConn *gorm.DB
	)

	BeforeEach(func() {
		// handy "hack" to create temp testible DB, open empty SQLite DB in memory
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		Expect(err).To(BeNil())
		Expect(db).ToNot(BeNil())
		mockDBConn = db

		models.AutoMigrate(mockDBConn)
	})

	AfterEach(func() {
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
				u := models.User{}
				err := repo.Create(&u)
				Expect(err).To(BeNil())
				existingUserUUID = u.UUID
			})

			It("Should find the existing user by it's UUID", func() {
				_, err := repo.FindByUUID(existingUserUUID)
				Expect(err).To(BeNil())
			})
		})
	})
})
