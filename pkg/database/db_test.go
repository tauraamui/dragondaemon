package data_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	data "github.com/tauraamui/dragondaemon/pkg/database"
	"github.com/tauraamui/dragondaemon/pkg/database/dbconn"
	"github.com/tauraamui/dragondaemon/pkg/database/models"
	"github.com/tauraamui/xerror"
)

type DBSetupTestSuite struct {
	suite.Suite
	dbMock                    dbconn.MockGormWrapper
	resetOpenDBConn           func()
	resetFs                   func()
	resetUC                   func()
	resetPlainPromptReader    func()
	resetPasswordPromptReader func()
}

func (suite *DBSetupTestSuite) SetupSuite() {
	suite.resetOpenDBConn = data.OverloadOpenDBConnection(
		func(string) (dbconn.GormWrapper, error) {
			return suite.dbMock, nil
		},
	)
}

func (suite *DBSetupTestSuite) TearDownSuite() {
	suite.resetOpenDBConn()
}

func (suite *DBSetupTestSuite) SetupTest() {
	suite.dbMock = dbconn.Mock()
	suite.resetFs = data.OverloadFS(afero.NewMemMapFs())
	suite.resetUC = data.OverloadUC(func() (string, error) {
		return "/testroot/.cache", nil
	})
	suite.resetPlainPromptReader = data.OverloadPlainPromptReader(
		testPlainPromptReader{
			testUsername: "testadmin",
		},
	)

	suite.resetPasswordPromptReader = data.OverloadPasswordPromptReader(
		testPasswordPromptReader{
			testPassword: "testpassword",
		},
	)
}

func (suite *DBSetupTestSuite) TearDownTest() {
	suite.dbMock = nil
	suite.resetFs()
	suite.resetUC()
	suite.resetPlainPromptReader()
	suite.resetPasswordPromptReader()
}

func (suite *DBSetupTestSuite) TestCreateFullFilePathForDBWithSingleRootUserDir() {
	is := is.New(suite.T())
	is.NoErr(data.Setup())

	created := suite.dbMock.Created()
	is.Equal(len(created), 1)
	user := models.User{}
	is.NoErr(dbconn.Replace(&user, created[0]))
	is.Equal(user.Name, "testadmin")
}

func (suite *DBSetupTestSuite) TestConnectWithoutHavingToRunSetupFirst() {
	is := is.New(suite.T())
	is.NoErr(data.Setup())

	conn, err := data.Connect()
	is.NoErr(err)
	is.True(conn != nil)
}

func (suite *DBSetupTestSuite) TestCreateFileAndThenRemovedOnDestroy() {
	is := is.New(suite.T())
	is.NoErr(data.Setup())

	is.NoErr(data.Destroy())

	is.Equal(data.Destroy().Error(), "remove /testroot/.cache/tacusci/dragondaemon/dd.db: file does not exist")
}

func (suite *DBSetupTestSuite) TestReturnErrorFromSetupDueToROFileSystem() {
	is := is.New(suite.T())
	suite.resetFs = data.OverloadFS(afero.NewReadOnlyFs(afero.NewMemMapFs()))
	is.Equal(data.Setup().Error(), "unable to create database file: operation not permitted")
}

func (suite *DBSetupTestSuite) TestReturnErrorFromSetupDueToDBAlreadyExisting() {
	is := is.New(suite.T())
	is.NoErr(data.Setup())
	is.Equal(data.Setup().Error(), "database file already exists: /testroot/.cache/tacusci/dragondaemon/dd.db")
}

func (suite *DBSetupTestSuite) TestReturnErrorFromSetupDueToPathResolutionFailure() {
	is := is.New(suite.T())
	suite.resetUC = data.OverloadUC(func() (string, error) {
		return "", xerror.New("test cache dir error")
	})
	is.Equal(data.Setup().Error(), "unable to resolve dd.db database file location: test cache dir error")
}

func (suite *DBSetupTestSuite) TestUnableToResolveDBPathHandlesAndReturnsWrappedError() {
	is := is.New(suite.T())
	is.NoErr(data.Setup())

	suite.resetUC = data.OverloadUC(func() (string, error) {
		return "", xerror.New("test cache dir error")
	})

	is.Equal(
		data.Destroy().Error(), "unable to delete database file: unable to resolve dd.db database file location: test cache dir error",
	)
}

func TestDBSetupTestSuite(t *testing.T) {
	suite.Run(t, &DBSetupTestSuite{})
}
