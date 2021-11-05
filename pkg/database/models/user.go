package models

import (
	"github.com/google/uuid"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/xerror"
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
		// TODO(tauraamui): change to return error instead
		log.Error(xerror.Errorf("unable to generate hash and salt from password: %w", err).Error()) //nolint
		return p
	}

	return string(h)
}

func cmp(h, p string) error {
	hb, pb := []byte(h), []byte(p)
	err := bcrypt.CompareHashAndPassword(hb, pb)
	if err != nil {
		return xerror.Errorf("incorrect password: %w", err)
	}

	return nil
}
