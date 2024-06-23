package transaction

import (
	"gorm.io/gorm"
)

type Repository interface {
	CreateTransaction(transaction Transaction) error
	GetTransactionByID(id string) (Transaction, error)
	UpdateTransactionStatus(id, status string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateTransaction(transaction Transaction) error {
	return r.db.Create(&transaction).Error
}

func (r *repository) GetTransactionByID(id string) (Transaction, error) {
	var transaction Transaction
	if err := r.db.Where("id = ?", id).First(&transaction).Error; err != nil {
		return Transaction{}, err
	}
	return transaction, nil
}

func (r *repository) UpdateTransactionStatus(id, status string) error {
	return r.db.Model(&Transaction{}).Where("id = ?", id).Update("status", status).Error
}
