package data_test

import (
	"github.com/matryer/is"
	"github.com/stretchr/testify/suite"
	data "github.com/tauraamui/dragondaemon/pkg/database"
	"github.com/tauraamui/dragondaemon/pkg/database/dbconn"
)

type DBSetupTestSuite struct {
	suite.Suite
	is              *is.I
	resetOpenDBConn func()
}

func (suite *DBSetupTestSuite) SetupSuite() {
	suite.resetOpenDBConn = data.OverloadOpenDBConnection(
		func(string) (dbconn.GormWrapper, error) {
			return nil, nil
		},
	)
}
