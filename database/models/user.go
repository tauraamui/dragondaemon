package models

import "gorm.io/gorm"

func init() {
	addToCollection(User{})
}

type User struct {
	gorm.Model
	Name     string
	AuthHash string
}
