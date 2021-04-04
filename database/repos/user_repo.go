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

func (r *UserRepository) Authenticate(username, password string) error {
	user := models.User{}
	if err := r.DB.Where("name = ?", username).First(&user).Error; err != nil {
		return fmt.Errorf("user of name %s not found", username)
	}

	return user.ComparePassword(password)
}
