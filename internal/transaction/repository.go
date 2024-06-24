package transaction

import (
	"context"
	"database/sql"
	"main/model"
	. "main/model"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	CreateTransaction(ctx context.Context, transaction Transaction) error
	GetTransactionByID(ctx context.Context, id string) (Transaction, error)
	UpdateTransactionStatus(ctx context.Context, id string, status model.TransactionStatus) error
	Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateTransaction(ctx context.Context, transaction Transaction) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	return r.db.WithContext(ctxTimeout).Create(&transaction).Error
}

func (r *repository) GetTransactionByID(ctx context.Context, id string) (Transaction, error) {
	var transaction Transaction
	if err := r.db.Where("transaction_id = ?", id).First(&transaction).Error; err != nil {
		return Transaction{}, err
	}
	return transaction, nil
}

func (r *repository) UpdateTransactionStatus(ctx context.Context, id string, status model.TransactionStatus) error {
	return r.db.Model(&Transaction{}).Where("transaction_id = ?", id).Update("transaction_status", status).Error
}

func (r *repository) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return r.db.Transaction(fc, opts...)
}
