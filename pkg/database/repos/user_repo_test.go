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

func (w *mockGormWrapper) First(dest interface{}, conds ...interface{}) (ref repos.GormWrapper) {
	ref = w
	if w.chain == nil {
		w.error = errors.New("need to call query first")
		return
	}

	w.chain.where.first = firstSelect{conds}
	w.error = replace(dest, w.result)

	return
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
	findFunc           func(string) (models.User, error)
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

	gorm := mockGormWrapper{result: existingUser}
	repo := repos.UserRepository{DB: &gorm}

	tests := []userRepoTest{
		{
			findFunc:           repo.FindByUUID,
			findWith:           "existing-test-user",
			expectedResultUUID: "existing-test-user",
			expectedResultName: "existing-test-user-name",
			expectedWhereQuery: "uuid = ?",
			expectedWhereArgs:  "existing-test-user",
		},
		{
			findFunc:           repo.FindByName,
			findWith:           "existing-test-user-name",
			expectedResultUUID: "existing-test-user",
			expectedResultName: "existing-test-user-name",
			expectedWhereQuery: "name = ?",
			expectedWhereArgs:  "existing-test-user-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			if tt.skip {
				t.Skip()
			}

			is := is.New(t)
			xis := xis.New(is)

			u, err := tt.findFunc(tt.findWith)
			is.NoErr(err)
			is.Equal(u.UUID, tt.expectedResultUUID)
			is.Equal(u.Name, tt.expectedResultName)

			is.Equal(gorm.chain.where.query, tt.expectedWhereQuery)
			xis.Contains(gorm.chain.where.args, tt.expectedWhereArgs)
			is.Equal(gorm.chain.where.first.conds, tt.expectedFirstConds)
		})
	}
}
