package account

import "time"

type FundMovementDirection string

var (
	Payment         FundMovementDirection = "Payment"
	PaymentReceived FundMovementDirection = "PaymentReceived"
)

type FundMovement struct {
	ID                   int       `gorm:"primaryKey;column:id" json:"id"`
	TransactionID        int       `gorm:"column:transaction_id" json:"transaction_id"`
	Direction            string    `gorm:"column:direction" json:"direction"`
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
