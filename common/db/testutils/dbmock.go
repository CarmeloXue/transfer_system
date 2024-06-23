package testutils

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetupTestDB() (*gorm.DB, error) {
	// Use an in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func PrepareData[T any](db *gorm.DB, data []T) {
	for _, acc := range data {
		result := db.Create(&acc)
		if result.Error != nil {
			panic("failed to seed data")
		}
	}
}
