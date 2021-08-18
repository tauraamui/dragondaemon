package config

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	fs afero.Fs
}

func (suite *ConfigTestSuite) SetupSuite() {
	suite.fs = afero.NewMemMapFs()
}

func (suite *ConfigTestSuite) TearDownSuite() {
	suite.fs = afero.NewMemMapFs()
}

func (suite *ConfigTestSuite) TestLoadConfig() {
	dir, err := resolveConfigPath()
	require.NoError(suite.T(), err)
	require.NoError(suite.T(), suite.fs.Mkdir(dir, os.ModeDir|os.ModePerm))
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, &ConfigTestSuite{})
}
