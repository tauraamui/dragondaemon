package repos

import "gorm.io/gorm"

type GormWrapper interface {
	Error() error
	Create(interface{}) GormWrapper
	Where(interface{}, ...interface{}) GormWrapper
	First(interface{}, ...interface{}) GormWrapper
}

type wrapper struct {
	db *gorm.DB
}

func Wrap(db *gorm.DB) GormWrapper {
	return &wrapper{
		db: db,
	}
}

func (w *wrapper) Error() error {
	return w.db.Error
}

func (w *wrapper) Create(value interface{}) GormWrapper {
	w.db.Create(value)
	return w
}

func (w *wrapper) Where(query interface{}, args ...interface{}) GormWrapper {
	w.db.Where(query, args...)
	return w
}

func (w *wrapper) First(dest interface{}, conds ...interface{}) GormWrapper {
	w.db.First(dest, conds)
	return w
}
