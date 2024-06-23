package transaction

import "time"

type Transaction struct {
	ID                   int       `gorm:"primaryKey;column:id" json:"id"`
	SourceAccountID      int       `gorm:"column:source_account_id" json:"source_account_id"`
	DestinationAccountID int       `gorm:"column:destination_account_id" json:"destination_account_id"`
	Amount               float64   `gorm:"column:amount" json:"amount"`
	Status               string    `gorm:"column:status" json:"status"`
	CreatedAt            time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName sets the insert table name for this struct type.
func (Transaction) TableName() string {
	return "transaction_tab"
}

type CreateTransactionRequest struct {
	SourceAccountID      string `json:"source_account_id" binding:"required"`
	DestinationAccountID string `json:"destination_account_id" binding:"required"`
	Amount               string `json:"amount" binding:"required"`
}

type ConfirmTransactionRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}
