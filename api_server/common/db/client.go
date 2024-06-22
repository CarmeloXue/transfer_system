package db

import (
	"errors"
	"fmt"
	"log"
	"main/models/account"
	"main/models/transaction"
	"os"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	accountDB   *gorm.DB
	accountOnce sync.Once

	transactionDB   *gorm.DB
	transactionOnce sync.Once
)

const (
	accountDBName     = "account_db"
	transactionDBName = "transaction_db"
)

func GetAccountDBClient() (*gorm.DB, error) {
	accountOnce.Do(func() {
		dbHost := os.Getenv("DATABASE_HOST")
		dbPort := os.Getenv("DATABASE_PORT")
		dbUser := os.Getenv("DATABASE_USER")
		dbPassword := os.Getenv("DATABASE_PASSWORD")
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPassword, accountDBName)

		var err error
		accountDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect database: %v", err)
		}
		_ = accountDB.AutoMigrate(account.Account{})
		_ = accountDB.AutoMigrate(account.FundMovement{})
	})
	if accountDB == nil {
		return nil, errors.New("failed to init database")
	}

	return accountDB, nil
}

func GetTransactionDB() (*gorm.DB, error) {
	transactionOnce.Do(func() {
		dbHost := os.Getenv("DATABASE_HOST")
		dbPort := os.Getenv("DATABASE_PORT")
		dbUser := os.Getenv("DATABASE_USER")
		dbPassword := os.Getenv("DATABASE_PASSWORD")
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPassword, transactionDBName)

		var err error
		transactionDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect database: %v", err)
		}
		_ = transactionDB.AutoMigrate(transaction.Transaction{})
	})
	if transactionDB == nil {
		return nil, errors.New("failed to init database")
	}

	return transactionDB, nil
}
