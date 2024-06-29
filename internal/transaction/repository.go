package transaction

import (
	"context"
	"database/sql"
	"errors"
	"main/common/config"
	"main/model"
	. "main/model"
	"time"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type Repository interface {
	CreateTransaction(ctx context.Context, transaction Transaction) error
	GetTransactionByID(ctx context.Context, id string) (Transaction, error)
	UpdateTransactionStatus(ctx context.Context, id string, status model.TransactionStatus) error
	Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error
	QueryExpiredTransactions(ctx context.Context) ([]model.Transaction, error)
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
	now := time.Now()
	// Set expiration time
	transaction.ExpiredAt = now.Add(time.Minute * time.Duration(viper.GetInt(config.ConfigKeyTransactionExpiration)))
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

func (r *repository) QueryExpiredTransactions(ctx context.Context) ([]model.Transaction, error) {
	var count int64
	if err := r.db.Model(Transaction{}).Where("expired_at < ?", time.Now()).Where("transaction_status in ?", []model.TransactionStatus{Pending, Processing}).Count(&count).Error; err != nil {
		return nil, err
	}

	if count > 200 {
		return nil, errors.New("too many pending/processing transactions")
	}

	var transactions []model.Transaction

	if err := r.db.Model(Transaction{}).Where("expired_at < ?", time.Now()).Where("transaction_status in ?", []model.TransactionStatus{Pending, Processing}).Offset(0).Limit(200).Find(&transactions).Error; err != nil {
		return nil, err
	}

	return transactions, nil
}
