package config

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
)

type CreateConfigTestSuite struct {
	suite.Suite
	configResolver configdef.Resolver
	fs             afero.Fs
}

func (suite *CreateConfigTestSuite) SetupSuite() {
	suite.fs = afero.NewMemMapFs()
	suite.configResolver = DefaultResolver()

	// use in memory FS in implementation for tests
	fs = suite.fs
}

func (suite *CreateConfigTestSuite) TearDownSuite() {
	suite.fs = afero.NewOsFs()
}

func (suite *CreateConfigTestSuite) TearDownTest() {
	suite.fs.RemoveAll("/")
}

func (suite *CreateConfigTestSuite) TestConfigCreate() {
	require.NoError(suite.T(), suite.configResolver.Create())
	loadedConfig, err := suite.configResolver.Resolve()

	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), configdef.Values{
		MaxClipAgeInDays: 30,
		Cameras:          []configdef.Camera{},
	}, loadedConfig)
}

func (suite *CreateConfigTestSuite) TestConfigCreateFailsDueToAlreadyExisting() {
	require.NoError(suite.T(), suite.configResolver.Create())
	err := suite.configResolver.Create()
	assert.EqualError(
		suite.T(), err,
		"config file already exists",
	)

	assert.ErrorIs(suite.T(), err, configdef.ErrConfigAlreadyExists)
}

func TestCreateConfigTestSuite(t *testing.T) {
	suite.Run(t, &CreateConfigTestSuite{})
}
