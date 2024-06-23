package model

import "time"

type FundMovementType string

var (
	FMPayment         FundMovementType = "Payment"
	FMPaymentReceived FundMovementType = "PaymentReceived"
	FMPaymentRefund   FundMovementType = "PaymentRefund" // This is used to refund a payment in cancel.
)

type FundMovement struct {
	ID                   int       `gorm:"primaryKey;column:id" json:"id"`
	TransactionID        string    `gorm:"column:transaction_id" json:"transaction_id"`
	FundMovementType     string    `gorm:"column:fund_movement_type" json:"fund_movement_type"`
	SourceAccountID      int       `gorm:"column:source_account_id" json:"source_account_id"`
	DestinationAccountID int       `gorm:"column:destination_account_id" json:"destination_account_id"`
	Amount               float64   `gorm:"column:amount" json:"amount"`
	CreatedAt            time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName sets the insert table name for this struct type.
func (FundMovement) TableName() string {
	return "fund_movement_tab"
}

type Account struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID int       `gorm:"unique;not null" json:"account_id"`
	Balance   float64   `gorm:"type:decimal(20,8);not null;default:0" json:"balance"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName sets the insert table name for this struct type.
func (Account) TableName() string {
	return "account_tab"
}
