package config

import (
	"os"
	"testing"

	"github.com/matryer/is"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/xerror"
)

type LoadConfigTestSuite struct {
	suite.Suite
	is             *is.I
	configResolver configdef.Resolver
	fs             afero.Fs
	path           string
	configFile     afero.File
}

func (suite *LoadConfigTestSuite) SetupSuite() {
	suite.fs = afero.NewMemMapFs()
	suite.configResolver = DefaultResolver()
	suite.is = is.New(suite.T())

	// use in memory FS in implementation for tests
	fs = suite.fs
}

func (suite *LoadConfigTestSuite) TearDownSuite() {
	suite.fs = afero.NewOsFs()
}

func (suite *LoadConfigTestSuite) SetupTest() {
	path, err := resolveConfigPath()
	suite.is.NoErr(err)
	suite.is.NoErr(suite.fs.MkdirAll(path, os.ModeDir|os.ModePerm))
	suite.path = path

	configFile, err := suite.fs.Create(path)
	suite.is.NoErr(err)
	suite.is.True(configFile != nil)

	suite.configFile = configFile

	// can be overridden this so reset it back before
	// each test to ensure that it's an opt in thing per
	// individual test
	suite.overwriteTestConfig(
		`{
			"debug": true,
			"secret": "DJIF3fje943fi4jefgo0",
			"cameras": []
		}`,
	)
}

func (suite *LoadConfigTestSuite) overwriteTestConfig(config string) {
	suite.is.NoErr(suite.configFile.Truncate(0))
	_, err := suite.configFile.Seek(0, 0)
	suite.is.NoErr(err)
	_, err = suite.configFile.WriteString(config)
	suite.is.NoErr(err)
}

func (suite *LoadConfigTestSuite) TearDownTest() {
	suite.is.NoErr(suite.configFile.Close())
	suite.is.NoErr(suite.fs.Remove(suite.path))
}

func (suite *LoadConfigTestSuite) TestLoadConfig() {
	config, err := suite.configResolver.Resolve()
	suite.is.NoErr(err)

	suite.is.True(config.Debug)
	suite.is.Equal(config.Secret, "DJIF3fje943fi4jefgo0")
	suite.is.Equal(config.Cameras, []configdef.Camera{})
}

func (suite *LoadConfigTestSuite) TestLoadConfigErrorOnResolvingUserConfigDir() {
	userConfigDirRef := userConfigDir
	userConfigDir = func() (string, error) {
		return "", xerror.New("test unable to resolve user config dir")
	}
	defer func() { userConfigDir = userConfigDirRef }()

	is := is.New(suite.T())

	_, err := suite.configResolver.Resolve()
	is.True(err != nil) // we need resolve to fail here
	is.Equal(
		err.Error(), "unable to resolve config.json location: test unable to resolve user config dir",
	)
}

func (suite *LoadConfigTestSuite) TestConfigLoadFailsValidationOnDupCameraTitles() {
	suite.overwriteTestConfig(
		`{"cameras": [
			{"title": "FakeCam1"},
			{"title": "FakeCam2"},
			{"title": "FakeCam3"},
			{"title": "FakeCam4"},
			{"title": "FakeCam3"}
		]}`,
	)

	config, err := suite.configResolver.Resolve()
	suite.is.True(err != nil)
	suite.is.Equal(config, configdef.Values{})
	suite.is.Equal(err.Error(), "validation failed: camera titles must be unique")
}

func TestLoadConfigTestSuite(t *testing.T) {
	suite.Run(t, &LoadConfigTestSuite{})
}
