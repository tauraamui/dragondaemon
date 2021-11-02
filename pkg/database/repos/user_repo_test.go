package repos_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/database/models"
	"github.com/tauraamui/dragondaemon/pkg/database/repos"
	"github.com/tauraamui/dragondaemon/pkg/xis"
)

type mockGormWrapper struct {
	error   error
	created []interface{}
	chain   *queryChain
	result  interface{}
}

type queryChain struct {
	where whereQuery
}

type whereQuery struct {
	query interface{}
	args  []interface{}
	first firstSelect
}

type firstSelect struct {
	conds []interface{}
}

func (w *mockGormWrapper) Error() error {
	return w.error
}

func (w *mockGormWrapper) Create(value interface{}) repos.GormWrapper {
	w.created = append(w.created, value)
	return w
}

func (w *mockGormWrapper) Where(query interface{}, args ...interface{}) repos.GormWrapper {
	w.chain = &queryChain{
		where: whereQuery{
			query: query,
			args:  args,
		},
	}
	return w
}

func (w *mockGormWrapper) First(dest interface{}, conds ...interface{}) repos.GormWrapper {
	if w.chain == nil {
		w.error = errors.New("need to call query first")
		return w
	}

	w.chain.where.first = firstSelect{conds}
	err := replace(dest, w.result)
	if w.error == nil {
		w.error = err
	}

	return w
}
func replace(i, v interface{}) error {
	val := reflect.ValueOf(i)
	if val.Kind() != reflect.Ptr {
		return errors.New("not a pointer")
	}

	val = val.Elem()

	newVal := reflect.Indirect(reflect.ValueOf(v))

	if !val.Type().AssignableTo(newVal.Type()) {
		return errors.New("mismatched types")
	}

	val.Set(newVal)
	return nil
}

type userRepoTest struct {
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

func TestUserRepo(t *testing.T) {
	existingUser := models.User{
		UUID: "existing-test-user",
		Name: "existing-test-user-name",
	}

	tests := []userRepoTest{
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
				UUID: "existing-test-user",
				Name: "existing-test-user-name",
			},
			findFunc:           "BYNAME",
			findWith:           "existing-test-user-name",
			expectedResultUUID: "existing-test-user",
			expectedResultName: "existing-test-user-name",
			expectedWhereQuery: "name = ?",
			expectedWhereArgs:  "existing-test-user-name",
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

			gorm := mockGormWrapper{result: existingUser, error: tt.error}
			repo := repos.UserRepository{DB: &gorm}
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

			is.Equal(gorm.chain.where.query, tt.expectedWhereQuery)
			xis.Contains(gorm.chain.where.args, tt.expectedWhereArgs)
			is.Equal(gorm.chain.where.first.conds, tt.expectedFirstConds)
		})
	}
}
