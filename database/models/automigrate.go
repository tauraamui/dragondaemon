package models

import "gorm.io/gorm"

type Model interface{}

var models = []Model{}

func addToCollection(m Model) {
	models = append(models, m)
}

func AutoMigrate(db *gorm.DB) error {
	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return err
		}
	}
	return nil
}
