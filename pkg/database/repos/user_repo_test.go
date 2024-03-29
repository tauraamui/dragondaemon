package repos_test

import (
	"errors"
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/database/dbconn"
	"github.com/tauraamui/dragondaemon/pkg/database/models"
	"github.com/tauraamui/dragondaemon/pkg/database/repos"
	"github.com/tauraamui/dragondaemon/pkg/xis"
)

func TestUserRepoCreateNoErr(t *testing.T) {
	is := is.New(t)
	xis := xis.New(is)

	gorm := dbconn.Mock()
	repo := repos.UserRepository{DB: gorm}

	user := models.User{
		Name: "new user",
	}
	is.NoErr(repo.Create(&user))
	xis.Contains(gorm.Created(), &user)
}

func TestUserRepoCreateWithErr(t *testing.T) {
	is := is.New(t)

	err := errors.New("unable to create data")
	gorm := dbconn.Mock().SetError(err)
	repo := repos.UserRepository{DB: gorm}

	user := models.User{
		Name: "new user",
	}
	is.Equal(repo.Create(&user).Error(), err.Error())
	is.Equal(len(gorm.Created()), 0)
}

type userRepoFindByTest struct {
	title              string
	skip               bool
	existingUser       models.User
	error              error
	findFunc           string
	findWith           string
	expectedResultUUID string
	expectedResultName string
	expectedWhereQuery string
	expectedWhereArgs  string
	expectedFirstConds []interface{}
}

func TestUserRepoFindBy(t *testing.T) {
	tests := []userRepoFindByTest{
		{
			title: "find user by uuid",
			existingUser: models.User{
				UUID: "existing-test-user",
				Name: "existing-test-user-name",
			},
			findFunc:           "BYUUID",
			findWith:           "existing-test-user",
			expectedResultUUID: "existing-test-user",
			expectedResultName: "existing-test-user-name",
			expectedWhereQuery: "uuid = ?",
			expectedWhereArgs:  "existing-test-user",
		},
		{
			title: "find user by name",
			existingUser: models.User{
				UUID: "existing-test-user-by-name",
				Name: "existing-test-user-slim-jim",
			},
			findFunc:           "BYNAME",
			findWith:           "existing-test-user-slim-jim",
			expectedResultUUID: "existing-test-user-by-name",
			expectedResultName: "existing-test-user-slim-jim",
			expectedWhereQuery: "name = ?",
			expectedWhereArgs:  "existing-test-user-slim-jim",
		},
		{
			title:    "find user by uuid returns error",
			findFunc: "BYUUID",
			findWith: "non-existent-uuid",
			error:    errors.New("user of uuid non-existent-uuid not found"),
		},
		{
			title:    "find user by name returns error",
			findFunc: "BYNAME",
			findWith: "non-existent-name",
			error:    errors.New("user of name non-existent-name not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			if tt.skip {
				t.Skip()
			}

			is := is.New(t)
			xis := xis.New(is)

			gorm := dbconn.Mock().SetResult(tt.existingUser).SetError(tt.error)
			repo := repos.UserRepository{DB: gorm}
			var findFunc func(string) (models.User, error)
			switch tt.findFunc {
			case "BYUUID":
				findFunc = repo.FindByUUID
			case "BYNAME":
				findFunc = repo.FindByName
			}

			u, err := findFunc(tt.findWith)
			if err != nil {
				is.Equal(err.Error(), tt.error.Error())
				return
			}

			is.Equal(u.UUID, tt.expectedResultUUID)
			is.Equal(u.Name, tt.expectedResultName)

			is.Equal(gorm.Chain().Where.Query, tt.expectedWhereQuery)
			xis.Contains(gorm.Chain().Where.Args, tt.expectedWhereArgs)
			is.Equal(gorm.Chain().Where.First.Conds, tt.expectedFirstConds)
		})
	}
}
