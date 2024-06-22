package account

import "gorm.io/gorm"

type AccountRepository interface {
	CreateAccount(account *Account) error
	GetAccountByID(id int) (Account, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) AccountRepository {
	return &repository{db: db}
}
