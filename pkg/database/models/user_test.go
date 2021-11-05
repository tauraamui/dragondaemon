package models_test

import (
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/database/models"
)

func TestEmptyUserBeforeCreateShouldGenerateUUIDAndEncryptAuthHash(t *testing.T) {
	is := is.New(t)
	user := models.User{}

	is.NoErr(user.BeforeCreate(nil))
	is.True(len(user.UUID) > 0)
}

func TestPopulatedUserBeforeCreateShouldGenerateUUIDAndEncryptAuthHash(t *testing.T) {
	is := is.New(t)
	user := models.User{
		Name:     "test-user-account",
		AuthHash: "test-user-password",
	}

	is.NoErr(user.BeforeCreate(nil))
	is.True(len(user.UUID) > 0)
	is.Equal(user.Name, "test-user-account")
	is.True(strings.Contains(user.AuthHash, "$2a$10$"))
	is.NoErr(user.ComparePassword("test-user-password")) // match enc auth hash to plaintxt password
}

func TestPopulatedUserBeforeCreateShouldFailToMatchPasswordIfIncorrect(t *testing.T) {
	is := is.New(t)
	user := models.User{
		Name:     "test-user-account",
		AuthHash: "test-user-password",
	}

	is.NoErr(user.BeforeCreate(nil))
	err := user.ComparePassword("incorrect-password")
	is.True(err != nil)
	is.Equal(err.Error(), "incorrect password: crypto/bcrypt: hashedPassword is not the hash of the given password") // fail to match enc auth hash to plaintxt password
}
