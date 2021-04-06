package repos

import (
	"fmt"

	"github.com/tauraamui/dragondaemon/database/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func (r *UserRepository) Create(user *models.User) error {
	return r.DB.Create(user).Error
}

func (r *UserRepository) FindUserByName(username string) (models.User, error) {
	user := models.User{}
	if err := r.DB.Where("name = ?", username).First(&user).Error; err != nil {
		return user, fmt.Errorf("user of name %s not found", username)
	}

	return user, nil
}
