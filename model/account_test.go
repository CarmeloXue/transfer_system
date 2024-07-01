package model_test

import (
	"main/model"
	"main/tools/currency"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock sql db, got error: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open gorm db, got error: %v", err)
	}

	return gormDB, mock
}

func TestAccount_TryTransfer(t *testing.T) {
	db, mock := setupMockDB(t)
	defer func() {
		db, err := db.DB()
		if err == nil {
			db.Close()
		}
	}()

	account := model.Account{
		AccountID:  1,
		Balance:    1000000,
		OutBalance: 200000,
	}
	amount := int64(300000)
	mock.ExpectBegin()
	// mock.ExpectCommit()
	// Mock expected query
	mock.ExpectExec(`UPDATE "account_tab".*`).
		WithArgs(500000, sqlmock.AnyArg(), account.AccountID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := account.TryTransfer(db, amount)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAccount_Transfer(t *testing.T) {
	db, _ := setupMockDB(t)
	defer func() {
		db, err := db.DB()
		if err == nil {
			db.Close()
		}
	}()

	account := model.Account{
		AccountID:  1,
		Balance:    1000000,
		OutBalance: 200000,
	}
	amount := int64(300000)

	err := account.Transfer(db, amount)
	assert.EqualError(t, currency.ErrNegativeValue, err.Error())
}

func TestAccount_TryReceive(t *testing.T) {
	db, mock := setupMockDB(t)
	defer func() {
		db, err := db.DB()
		if err == nil {
			db.Close()
		}
	}()
	account := model.Account{
		AccountID: 1,
		Balance:   1000000,
		InBalance: 200000,
	}
	amount := int64(300000)
	mock.ExpectBegin()
	// Mock expected query
	mock.ExpectExec(`UPDATE "account_tab".*`).
		WithArgs(500000, sqlmock.AnyArg(), account.AccountID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := account.TryReceive(db, amount)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAccount_Receive(t *testing.T) {
	db, mock := setupMockDB(t)
	defer func() {
		db, err := db.DB()
		if err == nil {
			db.Close()
		}
	}()
	account := model.Account{
		AccountID: 1,
		Balance:   1000000,
		InBalance: 300000,
	}
	amount := int64(300000)
	mock.ExpectBegin()
	// Mock expected query
	mock.ExpectExec(`UPDATE "account_tab".*`).
		WithArgs(1300000, 0, sqlmock.AnyArg(), account.AccountID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := account.Recieve(db, amount)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAccount_CancelTransfer(t *testing.T) {
	db, mock := setupMockDB(t)
	defer func() {
		db, err := db.DB()
		if err == nil {
			db.Close()
		}
	}()
	account := model.Account{
		AccountID:  1,
		Balance:    1000000,
		OutBalance: 300000,
	}

	amount := int64(300000)
	mock.ExpectBegin()
	// Mock expected query
	mock.ExpectExec(`UPDATE "account_tab".*`).
		WithArgs(0, sqlmock.AnyArg(), account.AccountID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := account.CancelTransfer(db, amount)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAccount_CancelRecieve(t *testing.T) {
	db, mock := setupMockDB(t)
	defer func() {
		db, err := db.DB()
		if err == nil {
			db.Close()
		}
	}()
	account := model.Account{
		AccountID: 1,
		Balance:   1000000,
		InBalance: 300000,
	}

	amount := int64(300000)
	mock.ExpectBegin()
	// Mock expected query
	mock.ExpectExec(`UPDATE "account_tab".*`).
		WithArgs(0, sqlmock.AnyArg(), account.AccountID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := account.CancelRecieve(db, amount)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
