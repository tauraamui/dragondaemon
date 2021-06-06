package models

import "gorm.io/gorm"

type Model interface{}

var models = []Model{}

func AutoMigrate(db *gorm.DB) error {
	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return err
		}
		m = nil
	}
	return nil
}

func registerForAutomigration(m Model) {
	models = append(models, m)
}
