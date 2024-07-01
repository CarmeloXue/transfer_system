package account

import (
	"context"
	"main/internal/common/config"
	"main/internal/common/db/testutils"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func init() {
	config.InitForTest()
}

func prepareRepo() (AccountRepository, error) {
	db, err := testutils.SetupTestDB()
	if err != nil {
		return nil, err
	}
	_ = db.AutoMigrate(Account{})
	return &repository{db}, nil
}

func TestCreateAccount(t *testing.T) {
	repo, err := prepareRepo()

	assert.NoError(t, err, "failed to connect mock db")

	// Your test logic here
	account := &Account{AccountID: 1, Balance: 100.0}

	err = repo.CreateAccount(context.Background(), account)
	assert.NoError(t, err, "failed to create account")

	count, err := repo.(*repository).countAccount(context.Background())
	assert.NoError(t, err, "failed to create account")

	assert.Equal(t, int64(1), count)
}

func TestCreateAccount_Duplicated_ShouldReturnError(t *testing.T) {
	repo, err := prepareRepo()

	assert.NoError(t, err, "failed to connect mock db")

	// Your test logic here
	account := &Account{AccountID: 1, Balance: 100.0}

	err = repo.CreateAccount(context.Background(), account)
	assert.NoError(t, err, "failed to create account")

	err = repo.CreateAccount(context.Background(), account)
	assert.EqualError(t, ErrInternalDuplicatedAccount, err.Error())

	count, err := repo.(*repository).countAccount(context.Background())
	assert.NoError(t, err, "failed to create account")

	assert.Equal(t, int64(1), count)
}

func TestGetAccount(t *testing.T) {
	repo, err := prepareRepo()
	assert.NoError(t, err, "failed to test")

	account := &Account{AccountID: 123, Balance: 100}

	err = repo.CreateAccount(context.Background(), account)
	assert.NoError(t, err, "failed to create account")
	acc, err := repo.GetAccountByID(context.Background(), 123)
	assert.NoError(t, err, "failed to create account")

	assert.Equal(t, 123, acc.AccountID)
	assert.Equal(t, int64(100), acc.Balance)
}

func TestGetAccount_NoData(t *testing.T) {
	repo, err := prepareRepo()
	assert.NoError(t, err, "failed to test")

	_, err = repo.GetAccountByID(context.Background(), 1)
	assert.Error(t, err, "failed to create account")
	assert.EqualError(t, gorm.ErrRecordNotFound, err.Error())
}
