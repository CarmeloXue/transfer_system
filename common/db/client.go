package db

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

func GetDBClient() (*gorm.DB, error) {
	once.Do(func() {
		dbHost := os.Getenv("DATABASE_HOST")
		dbPort := os.Getenv("DATABASE_PORT")
		dbUser := os.Getenv("DATABASE_USER")
		dbPassword := os.Getenv("DATABASE_PASSWORD")
		dbName := os.Getenv("DATABASE_NAME")
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPassword, dbName)

		var err error
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect database: %v", err)
		}
	})
	if db == nil {
		return nil, errors.New("failed to init database")
	}
	return db, nil
}
