package config

import (
	"errors"
	"testing"

	"github.com/matryer/is"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
)

type CreateConfigTestSuite struct {
	suite.Suite
	is                   *is.I
	configCreateResolver configdef.CreateResolver
	fs                   afero.Fs
}

func (suite *CreateConfigTestSuite) SetupSuite() {
	suite.is = is.New(suite.T())
	suite.fs = afero.NewMemMapFs()
	suite.configCreateResolver = DefaultCreateResolver()

	// use in memory FS in implementation for tests
	fs = suite.fs
}

func (suite *CreateConfigTestSuite) TearDownSuite() {
	suite.fs = afero.NewOsFs()
}

func (suite *CreateConfigTestSuite) TearDownTest() {
	suite.is.NoErr(suite.fs.RemoveAll("/"))
}

func (suite *CreateConfigTestSuite) TestConfigCreate() {
	require.NoError(suite.T(), suite.configCreateResolver.Create())
	loadedConfig, err := suite.configCreateResolver.Resolve()

	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), configdef.Values{
		Cameras: []configdef.Camera{},
	}, loadedConfig)
}

func (suite *CreateConfigTestSuite) TestConfigCreateFailsDueToAlreadyExisting() {
	suite.is.NoErr(suite.configCreateResolver.Create())
	err := suite.configCreateResolver.Create()
	suite.is.Equal(err.Error(), "config file already exists")
	suite.is.True(errors.Is(err, configdef.ErrConfigAlreadyExists))
}

func TestCreateConfigTestSuite(t *testing.T) {
	suite.Run(t, &CreateConfigTestSuite{})
}
