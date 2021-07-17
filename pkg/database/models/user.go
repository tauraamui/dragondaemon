package models

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func init() {
	registerForAutomigration(&User{})
}

type User struct {
	gorm.Model
	UUID         string
	Name         string
	AuthHash     string
	SessionToken string
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.UUID = uuid.NewString()
	u.AuthHash = enc(u.AuthHash)
	return nil
}

func (u *User) ComparePassword(password string) error {
	return cmp(u.AuthHash, password)
}

func enc(p string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		log.Error(fmt.Errorf("unable to generate hash and salt from password: %w", err).Error()) //nolint
		return p
	}

	return string(h)
}

func cmp(h, p string) error {
	hb, pb := []byte(h), []byte(p)
	err := bcrypt.CompareHashAndPassword(hb, pb)
	if err != nil {
		return fmt.Errorf("password does not match hash: %w", err)
	}

	return nil
}
