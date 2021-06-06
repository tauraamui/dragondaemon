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

type testReader struct {
	onReadCallback func()
	readData       []byte
	readError      error
}

func (t testReader) Read(b []byte) (int, error) {
	t.onReadCallback()
	n := copy(b, t.readData)
	return n, t.readError
}

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
	attemptCount, maxCalls int
	passwordsToAttempt     []string
	testError              error
}

func (t *multipleAttemptPasswordPromptReader) ReadPassword(string) ([]byte, error) {
	if t.attemptCount >= t.maxCalls {
		return nil, errors.New("TESTING ERROR: multipleAttempts exceeds maximum call limit")
	}
	password := []byte(t.passwordsToAttempt[t.attemptCount])
	t.attemptCount++
	return password, t.testError
}

var _ = Describe("Data", func() {
	existingLoggingLevel := logging.CurrentLoggingLevel

	var resetFs func() = nil

	BeforeEach(func() {
		logging.CurrentLoggingLevel = logging.SilentLevel
		resetFs = data.OverloadFS(afero.NewMemMapFs())
	})

	AfterEach(func() {
		logging.CurrentLoggingLevel = existingLoggingLevel
		resetFs()
	})

	Context("Running setup", func() {
		var resetFs func() = nil
		var resetPlainPromptReader func()
		var resetPasswordPromptReader func()

		BeforeEach(func() {
			resetFs = data.OverloadFS(afero.NewMemMapFs())
			resetPlainPromptReader = data.OverloadPlainPromptReader(
				testPlainPromptReader{
					testUsername: "testadmin",
				},
			)

			resetPasswordPromptReader = data.OverloadPasswordPromptReader(
				testPasswordPromptReader{
					testPassword: "testpassword",
				},
			)
		})

		AfterEach(func() {
			resetFs()
			resetPlainPromptReader()
			resetPasswordPromptReader()
		})

		It("Should create full file path for DB with single root user entry", func() {
			err := data.Setup()
			Expect(err).To(BeNil())

			conn, err := data.Connect()
			Expect(err).To(BeNil())
			userRepo := repos.UserRepository{DB: conn}

			user, err := userRepo.FindByName("testadmin")
			Expect(err).To(BeNil())
			Expect(user.Name).To(Equal("testadmin"))
		})

		It("Should create file and then be removed on destroy call", func() {
			err := data.Setup()
			Expect(err).To(BeNil())

			err = data.Destroy()
			Expect(err).To(BeNil())

			err = data.Destroy()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("remove /home/tauraamui/.cache/tacusci/dragondaemon/dd.db: no such file or directory"))
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

		It("Should handle unable to resolve DB path gracefully and return wrapped error", func() {
			err := data.Setup()
			Expect(err).To(BeNil())

			reset := data.OverloadUC(func() (string, error) {
				return "", errors.New("test cache dir error")
			})
			defer reset()

			err = data.Destroy()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(
				"unable to delete database file: unable to resolve dd.db database file location: test cache dir error",
			))
		})

		Context("Reading new root username and password input", func() {
			It("Should handle username prompt error gracefully and return wrapped error", func() {
				resetPlainPromptReader := data.OverloadPlainPromptReader(
					testPlainPromptReader{
						testError: errors.New("testing read username error"),
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

				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("failed to prompt for root username: testing read username error"))
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
						maxCalls: 6,
						passwordsToAttempt: []string{
							"1stpair", "1stpairnomatch", "2ndpair", "2ndpairnomatch", "3rdpair", "3rdpairnomatch",
						},
					},
				)
				defer resetPasswordPromptReader()

				err := data.Setup()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("failed to prompt for root password: tried entering new password at least 3 times"))
			})
		})
	})

	Context("Plain prompt reader implementation", func() {
		It("Should read from given readable and return value", func() {
			calledCount := 0
			plainReader := data.NewStdinPlainReader(
				testReader{
					readData: []byte("testuser\n"),
					onReadCallback: func() {
						calledCount++
					},
				},
			)

			value, err := plainReader.ReadPlain("")
			Expect(err).To(BeNil())
			Expect(value).To(Equal("testuser"))
			Expect(calledCount).To(Equal(1))
		})
	})
})
