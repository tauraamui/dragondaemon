package models

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tacusci/logging/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func init() {
	addToCollection(&User{})
}

type User struct {
	gorm.Model
	UUID     string
	Name     string
	AuthHash string
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.UUID = uuid.NewString()
	u.AuthHash = enc(u.AuthHash)
	return nil
}

func enc(p string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		logging.Error(fmt.Errorf("unable to generate hash and salt from password: %w", err).Error())
		return p
	}

	return string(h)
}

func cmp(h, p string) bool {
	hb, pb := []byte(h), []byte(p)
	err := bcrypt.CompareHashAndPassword(hb, pb)
	if err != nil {
		logging.Error(fmt.Errorf("comparing hashed password to plain password failed: %w", err).Error())
		return false
	}

	return true
}
