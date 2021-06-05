package data_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/tacusci/logging/v2"
	data "github.com/tauraamui/dragondaemon/pkg/database"
	"github.com/tauraamui/dragondaemon/pkg/database/repos"
)

type testPlainPromptReader struct {
	testUsername string
	testError    error
}

func (t testPlainPromptReader) ReadPlain(string) (string, error) {
	return t.testUsername, t.testError
}

type testPasswordPromptReader struct {
	testPassword string
	testError    error
}

func (t testPasswordPromptReader) ReadPassword(string) ([]byte, error) {
	return []byte(t.testPassword), t.testError
}

type multipleAttemptPasswordPromptReader struct {
	attemptCount       int
	passwordsToAttempt []string
	testError          error
}

func (t *multipleAttemptPasswordPromptReader) ReadPassword(string) ([]byte, error) {
	password := []byte(t.passwordsToAttempt[t.attemptCount])
	t.attemptCount++
	return password, t.testError
}

var _ = Describe("Data", func() {
	existingLoggingLevel := logging.CurrentLoggingLevel

	BeforeEach(func() {
		logging.CurrentLoggingLevel = logging.SilentLevel
	})

	AfterEach(func() {
		logging.CurrentLoggingLevel = existingLoggingLevel
	})

	Context("Setup run against blank file system", func() {
		var resetFs func() = nil

		BeforeEach(func() {
			resetFs = data.OverloadFS(afero.NewMemMapFs())
		})

		AfterEach(func() {
			resetFs()
		})

		It("Should create full file path for DB with single root user entry", func() {
			resetPlainPromptReader := data.OverloadPlainPromptReader(
				testPlainPromptReader{
					testUsername: "testadmin",
				},
			)
			defer resetPlainPromptReader()

			resetPasswordPromptReader := data.OverloadPasswordPromptReader(
				testPasswordPromptReader{
					testPassword: "testpassword",
				},
			)
			defer resetPasswordPromptReader()

			err := data.Setup()
			Expect(err).To(BeNil())

			conn, err := data.Connect()
			Expect(err).To(BeNil())
			userRepo := repos.UserRepository{DB: conn}

			user, err := userRepo.FindByName("testadmin")
			Expect(err).To(BeNil())
			Expect(user.Name).To(Equal("testadmin"))
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

		It("Should return error from too many incorrect password attempts", func() {
			resetPlainPromptReader := data.OverloadPlainPromptReader(
				testPlainPromptReader{
					testUsername: "testadmin",
				},
			)
			defer resetPlainPromptReader()

			resetPasswordPromptReader := data.OverloadPasswordPromptReader(
				&multipleAttemptPasswordPromptReader{
					attemptCount:       0,
					passwordsToAttempt: []string{"actual", "firstrepeat", "secondrepeat", "thirdrepeat"},
				},
			)
			defer resetPasswordPromptReader()

			err := data.Setup()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("failed to prompt for root password: tried entering new password at least 3 times"))
		})
	})
})
