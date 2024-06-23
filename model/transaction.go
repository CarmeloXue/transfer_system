package model

import "time"

type TransactionStatus string

var (
	Pending    TransactionStatus = "pending"
	Processing TransactionStatus = "processing"
	Failed     TransactionStatus = "failed"
	Fulfiled   TransactionStatus = "fulfiled"
	Refunded   TransactionStatus = "refunded"
)

type Transaction struct {
	ID                   uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TransactionID        string    `gorm:"unique;not null" json:"transaction_id"`
	SourceAccountID      int       `gorm:"not null" json:"source_account_id"`
	DestinationAccountID int       `gorm:"not null" json:"destination_account_id"`
	Amount               float64   `gorm:"type:decimal(20,8);not null" json:"amount"`
	Status               string    `gorm:"type:varchar(20);not null" json:"status"`
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Retries              int       `gorm:"-"`
}

// TableName sets the insert table name for this struct type.
func (Transaction) TableName() string {
	return "transaction_tab"
}
