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
	if w.chain == nil {
		w.chain = &queryChain{
			where: whereQuery{
				query: query,
				args:  args,
			},
		}
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

func TestUserRepoFindUserByUUID(t *testing.T) {
	is := is.New(t)
	xis := xis.New(is)

	existingUser := models.User{
		UUID: "existing-test-user",
		Name: "existing-test-user-name",
	}

	gorm := mockGormWrapper{result: existingUser}
	repo := repos.UserRepository{DB: &gorm}

	u, err := repo.FindByUUID("existing-test-user")
	is.NoErr(err)
	is.Equal(u.Name, "existing-test-user-name")

	is.Equal(gorm.chain.where.query, "uuid = ?")
	xis.Contains(gorm.chain.where.args, "existing-test-user")
	is.Equal(len(gorm.chain.where.first.conds), 0)
}

func TestUserRepoFindUserByName(t *testing.T) {
	is := is.New(t)
	xis := xis.New(is)

	existingUser := models.User{
		UUID: "existing-test-user",
		Name: "existing-test-user-name",
	}

	gorm := mockGormWrapper{result: existingUser}
	repo := repos.UserRepository{DB: &gorm}

	u, err := repo.FindByName("existing-test-user-name")
	is.NoErr(err)
	is.Equal(u.UUID, "existing-test-user")

	is.Equal(gorm.chain.where.query, "name = ?")
	xis.Contains(gorm.chain.where.args, "existing-test-user-name")
	is.Equal(len(gorm.chain.where.first.conds), 0)
}
