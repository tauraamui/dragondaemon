package config

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
)

type LoadConfigTestSuite struct {
	suite.Suite
	configResolver configdef.Resolver
	fs             afero.Fs
	path           string
	configFile     afero.File
}

func (suite *LoadConfigTestSuite) SetupSuite() {
	suite.fs = afero.NewMemMapFs()
	suite.configResolver = DefaultResolver()

	// use in memory FS in implementation for tests
	fs = suite.fs
}

func (suite *LoadConfigTestSuite) TearDownSuite() {
	suite.fs = afero.NewOsFs()
}

func (suite *LoadConfigTestSuite) SetupTest() {
	path, err := resolveConfigPath()
	require.NoError(suite.T(), err)
	require.NoError(suite.T(), suite.fs.MkdirAll(path, os.ModeDir|os.ModePerm))
	suite.path = path

	configFile, err := suite.fs.Create(path)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), configFile)

	suite.configFile = configFile

	// can be overridden this so reset it back before
	// each test to ensure that it's an opt in thing per
	// individual test
	suite.overwriteTestConfig(
		`{
			"debug": true,
			"secret": "DJIF3fje943fi4jefgo0",
			"max_clip_age_in_days": 19,
			"cameras": []
		}`,
	)
}

func (suite *LoadConfigTestSuite) overwriteTestConfig(config string) {
	require.NoError(suite.T(), suite.configFile.Truncate(0))
	_, err := suite.configFile.Seek(0, 0)
	require.NoError(suite.T(), err)
	_, err = suite.configFile.WriteString(config)
	assert.NoError(suite.T(), err)
}

func (suite *LoadConfigTestSuite) TearDownTest() {
	require.NoError(suite.T(), suite.configFile.Close())
	suite.fs.Remove(suite.path)
}

func (suite *LoadConfigTestSuite) TestLoadConfig() {
	config, err := suite.configResolver.Resolve()
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), config)

	assert.Equal(suite.T(), true, config.Debug)
	assert.Equal(suite.T(), "DJIF3fje943fi4jefgo0", config.Secret)
	assert.Equal(suite.T(), 19, config.MaxClipAgeInDays)
	assert.ElementsMatch(suite.T(), config.Cameras, []configdef.Camera{})
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
	require.Error(suite.T(), err)
	require.Empty(suite.T(), config)

	assert.EqualError(suite.T(), err, "validation failed: camera titles must be unique")
}

func TestLoadConfigTestSuite(t *testing.T) {
	suite.Run(t, &LoadConfigTestSuite{})
}
