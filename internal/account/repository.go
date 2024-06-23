package account

import (
	"context"

	. "main/model"

	"gorm.io/gorm"
)

// TODO: Add timeout in implementation
type AccountRepository interface {
	CreateAccount(ctx context.Context, account *Account) error
	GetAccountByID(ctx context.Context, id int) (Account, error)

	GetFundMovement(ctx context.Context, query FundMovement) (*FundMovement, error)
	QueryFundMovement(ctx context.Context, transactionID string) ([]FundMovement, error)
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

func (r *repository) GetFundMovement(_ context.Context, query FundMovement) (*FundMovement, error) {
	var fm FundMovement
	if err := r.db.First(&fm, query).Error; err != nil {
		return nil, err
	}

	return &fm, nil
}

func (r *repository) QueryFundMovement(_ context.Context, transactionID string) ([]FundMovement, error) {
	var fundmvmts = make([]FundMovement, 3) // there are only 2(I'm thinking to add a refund type for mannual refund) and trx_id - fund_movement_type is unique key

	if err := r.db.Where("transaction_id = ?", transactionID).Find(&fundmvmts).Error; err != nil {
		return nil, err
	}

	return fundmvmts, nil
}
