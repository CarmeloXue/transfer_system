package account

import (
	"context"

	"gorm.io/gorm"
)

// TODO: Add timeout in implementation
type AccountRepository interface {
	CreateAccount(ctx context.Context, account *Account) error
	GetAccountByID(ctx context.Context, id int) (Account, error)

	GetFundMovement(ctx context.Context, transactionID, sourceID int, fundMovementType FundMovementType) (FundMovement, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) AccountRepository {
	return &repository{db: db}
}

func (r *repository) CreateAccount(_ context.Context, account *Account) error {
	if err := r.db.Create(account).Error; err != nil {
		return err
	}
	return nil
}

func (r *repository) GetAccountByID(_ context.Context, accountID int) (Account, error) {
	var acc Account
	if err := r.db.First(&acc, Account{
		AccountID: accountID,
	}).Error; err != nil {
		return Account{}, err
	}
	return acc, nil
}

func (r *repository) countAccount(_ context.Context) (int64, error) {
	var count int64
	if err := r.db.Model(Account{}).Count(&count).Error; err != nil {
		return count, err
	}
	return count, nil
}

func (r *repository) GetFundMovement(_ context.Context, transactionID, sourceID int, fundMovementType FundMovementType) (FundMovement, error) {
	var fm *FundMovement
	r.db.First(fm, FundMovement{
		TransactionID:    transactionID,
		SourceAccountID:  sourceID,
		FundMovementType: string(fundMovementType),
	})
	return FundMovement{}, nil
}
