package data_test

import (
	"errors"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/tacusci/logging/v2"
	data "github.com/tauraamui/dragondaemon/pkg/database"
	"github.com/tauraamui/dragondaemon/pkg/database/repos"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	var resetUC func() = nil

	BeforeEach(func() {
		logging.CurrentLoggingLevel = logging.SilentLevel
		resetFs = data.OverloadFS(afero.NewMemMapFs())
		resetUC = data.OverloadUC(func() (string, error) {
			return "/testroot/.cache", nil
		})
	})

	AfterEach(func() {
		logging.CurrentLoggingLevel = existingLoggingLevel
		resetFs()
		resetUC()
	})

	Context("Running setup", func() {
		var resetFs func() = nil
		var mockFs afero.Fs = nil
		var resetOpenDBConn func()
		var resetPlainPromptReader func()
		var resetPasswordPromptReader func()

		BeforeSuite(func() {
			resetOpenDBConn = data.OverloadOpenDBConnection(
				func(string) (*gorm.DB, error) {
					return gorm.Open(sqlite.Open("file::memory:?cache=shared"))
				},
			)
		})

		BeforeEach(func() {
			mockFs = afero.NewMemMapFs()
			resetFs = data.OverloadFS(mockFs)
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

		JustBeforeEach(func() {
			Expect(mockFs.MkdirAll("/testroot/.cache", os.ModeDir|os.ModePerm)).To(BeNil())
		})

		AfterEach(func() {
			resetFs()
			resetPlainPromptReader()
			resetPasswordPromptReader()
			mockFs = nil
		})

		AfterSuite(func() {
			resetOpenDBConn()
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

		It("Should connect without having to run setup first", func() {
			err := data.Setup()
			Expect(err).To(BeNil())

			conn, err := data.Connect()
			Expect(err).To(BeNil())

			Expect(conn).ToNot(BeNil())
		})

		It("Should create file and then be removed on destroy call", func() {
			err := data.Setup()
			Expect(err).To(BeNil())

			err = data.Destroy()
			Expect(err).To(BeNil())

			err = data.Destroy()
			Expect(err).To(MatchError("remove /testroot/.cache/tacusci/dragondaemon/dd.db: file does not exist"))
		})

		It("Should return error from setup due to read only fs", func() {
			resetFs = data.OverloadFS(afero.NewReadOnlyFs(afero.NewMemMapFs()))
			err := data.Setup()
			Expect(err).To(MatchError("unable to create database file: operation not permitted"))
		})

		It("Should return error from setup due to db already existing", func() {
			err := data.Setup()
			Expect(err).To(BeNil())

			err = data.Setup()
			Expect(err).To(MatchError("database file already exists: /testroot/.cache/tacusci/dragondaemon/dd.db"))
		})

		It("Should return error from setup due to path resolution failure", func() {
			resetUC = data.OverloadUC(func() (string, error) {
				return "", errors.New("test cache dir error")
			})

			err := data.Setup()

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unable to resolve dd.db database file location: test cache dir error"))
		})

		It("Should handle unable to resolve DB path gracefully and return wrapped error", func() {
			err := data.Setup()
			Expect(err).To(BeNil())

			resetUC = data.OverloadUC(func() (string, error) {
				return "", errors.New("test cache dir error")
			})

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
