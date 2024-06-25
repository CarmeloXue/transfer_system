package model

import (
	"main/common/utils"
	"time"
)

type TransactionStatus int

var (
	Pending    TransactionStatus = 1
	Processing TransactionStatus = 2
	Fulfiled   TransactionStatus = 3
	Failed     TransactionStatus = 5
)

type Transaction struct {
	ID                   uint              `gorm:"primaryKey;autoIncrement" json:"-"`
	TransactionID        string            `gorm:"unique;not null" json:"transaction_id"`
	SourceAccountID      int               `gorm:"not null" json:"source_account_id"`
	DestinationAccountID int               `gorm:"not null" json:"destination_account_id"`
	Amount               int64             `gorm:"type:decimal(20,8);not null" json:"amount,omitempty"`
	TransactionStatus    TransactionStatus `gorm:"type:int;not null" json:"transaction_status"`
	CreatedAt            time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
	Retries              int               `gorm:"-" json:"-"`
	TransactionAmount    string            `gorm:"_" json:"transaction_amount,omitempty"`
}

// TableName sets the insert table name for this struct type.
func (Transaction) TableName() string {
	return "transaction_tab"
}

func (t *Transaction) FormatForDisplay() {
	t.TransactionAmount = utils.FormatInt(t.Amount)
	t.Amount = 0
}
