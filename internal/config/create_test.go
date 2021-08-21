package config

import (
	"testing"

	"github.com/spf13/afero"
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

func (suite *CreateConfigTestSuite) TestConfigCreate() {
	require.NoError(suite.T(), suite.configResolver.Create())
}

func TestCreateConfigTestSuite(t *testing.T) {
	suite.Run(t, &CreateConfigTestSuite{})
}
