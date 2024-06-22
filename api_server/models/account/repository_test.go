package account

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() (*gorm.DB, error) {
	// Use an in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate the Account struct to create the schema
	db.AutoMigrate(&Account{})

	return db, nil
}

func prepareRepo() (AccountRepository, error) {
	db, err := setupTestDB()
	if err != nil {
		return nil, err
	}
	return &repository{db}, nil
}

func TestCreateAccount(t *testing.T) {
	repo, err := prepareRepo()

	assert.NoError(t, err, "failed to connect mock db")

	// Your test logic here
	account := &Account{AccountID: 1, Balance: 100.0}

	err = repo.CreateAccount(account)
	assert.NoError(t, err, "failed to create account")

	count, err := repo.(*repository).countAccount()
	assert.NoError(t, err, "failed to create account")

	assert.Equal(t, int64(1), count)
}

func TestGetAccount(t *testing.T) {
	repo, err := prepareRepo()
	assert.NoError(t, err, "failed to test")

	// Your test logic here
	account := &Account{AccountID: 123, Balance: 100.0}

	err = repo.CreateAccount(account)
	assert.NoError(t, err, "failed to create account")
	acc, err := repo.GetAccountByID(1)
	assert.NoError(t, err, "failed to create account")

	assert.Equal(t, 123, acc.AccountID)
	assert.Equal(t, 100.0, acc.Balance)
}

func TestGetAccount_NoData(t *testing.T) {
	repo, err := prepareRepo()
	assert.NoError(t, err, "failed to test")

	_, err = repo.GetAccountByID(1)
	assert.Error(t, err, "failed to create account")
	assert.EqualError(t, gorm.ErrRecordNotFound, err.Error())
}
