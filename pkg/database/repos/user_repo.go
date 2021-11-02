package repos

import (
	"github.com/tauraamui/dragondaemon/pkg/database/models"
	"github.com/tauraamui/xerror"
)

type UserRepository struct {
	DB GormWrapper
}

func (r *UserRepository) Create(user *models.User) error {
	return r.DB.Create(user).Error()
}

func (r *UserRepository) FindByUUID(uuid string) (models.User, error) {
	user := models.User{}
	if err := r.DB.Where("uuid = ?", uuid).First(&user).Error(); err != nil {
		return user, xerror.Errorf("user of uuid %s not found", uuid)
	}

	return user, nil
}

func (r *UserRepository) FindByName(username string) (models.User, error) {
	user := models.User{}
	if err := r.DB.Where("name = ?", username).First(&user).Error(); err != nil {
		return user, xerror.Errorf("user of name %s not found", username)
	}

	return user, nil
}
