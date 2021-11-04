package dbconn

import (
	"errors"
	"reflect"
)

type MockGormWrapper interface {
	GormWrapper
	Created() []interface{}
	Chain() *queryChain
	SetError(error) MockGormWrapper
	SetResult(interface{}) MockGormWrapper
}

type mockGormWrapper struct {
	error   error
	created []interface{}
	chain   *queryChain
	result  interface{}
}

type queryChain struct {
	Where whereQuery
}

type whereQuery struct {
	Query interface{}
	Args  []interface{}
	First firstSelect
}

type firstSelect struct {
	Conds []interface{}
}

func Mock() MockGormWrapper {
	return &mockGormWrapper{}
}

func (w *mockGormWrapper) Created() []interface{} {
	return w.created
}

func (w *mockGormWrapper) Chain() *queryChain {
	return w.chain
}

func (w *mockGormWrapper) SetError(e error) MockGormWrapper {
	w.error = e
	return w
}

func (w *mockGormWrapper) SetResult(r interface{}) MockGormWrapper {
	w.result = r
	return w
}

func (w *mockGormWrapper) Error() error {
	return w.error
}

func (w *mockGormWrapper) AutoMigrate(...interface{}) error {
	return nil
}

func (w *mockGormWrapper) Create(value interface{}) GormWrapper {
	if w.error == nil {
		w.created = append(w.created, value)
	}
	return w
}

func (w *mockGormWrapper) Where(query interface{}, args ...interface{}) GormWrapper {
	w.chain = &queryChain{
		Where: whereQuery{
			Query: query,
			Args:  args,
		},
	}
	return w
}

func (w *mockGormWrapper) First(dest interface{}, conds ...interface{}) GormWrapper {
	if w.chain == nil {
		w.error = errors.New("need to call query first")
		return w
	}

	w.chain.Where.First = firstSelect{conds}
	err := Replace(dest, w.result)
	if w.error == nil {
		w.error = err
	}

	return w
}
func Replace(i, v interface{}) error {
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
