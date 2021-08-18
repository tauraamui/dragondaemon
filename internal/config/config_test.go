package config

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	fs          afero.Fs
	fileContent string
}

func (suite *ConfigTestSuite) SetupSuite() {
	suite.fs = afero.NewMemMapFs()
	suite.fileContent = `
		{
			"max_clip_age_in_days": 30,
			"cameras": []
		}
	`
}

func (suite *ConfigTestSuite) TearDownSuite() {
	suite.fs = afero.NewMemMapFs()
}

func (suite *ConfigTestSuite) TestLoadConfig() {
	path, err := resolveConfigPath()
	require.NoError(suite.T(), err)
	require.NoError(suite.T(), suite.fs.MkdirAll(path, os.ModeDir|os.ModePerm))

	configFile, err := suite.fs.Create(path)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), configFile)

	_, err = configFile.WriteString(suite.fileContent)
	assert.NoError(suite.T(), err)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, &ConfigTestSuite{})
}
