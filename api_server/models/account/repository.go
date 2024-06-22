package account

import (
	"context"
	"main/models/transaction"

	"gorm.io/gorm"
)

type AccountRepository interface {
	CreateAccount(account *Account) error
	GetAccountByID(id int) (Account, error)

	// TCC api
	Try(ctx context.Context, trx transaction.Transaction) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) AccountRepository {
	return &repository{db: db}
}
