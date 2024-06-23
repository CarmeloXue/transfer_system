package account

import "time"

type FundMovementType string

var (
	Payment         FundMovementType = "Payment"
	PaymentReceived FundMovementType = "PaymentReceived"
)

type FundMovement struct {
	ID                   int       `gorm:"primaryKey;column:id" json:"id"`
	TransactionID        int       `gorm:"column:transaction_id" json:"transaction_id"`
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

type (
	CreateAccountRequest struct {
		AccountID      uint64 `json:"account_id" binding:"required"`
		InitialBalance string `json:"initial_balance"`
	}

	// CreateAccountResponse represents the JSON response body structure
	CreateAccountResponse struct {
		AccountID uint64 `json:"account_id"`
		Balance   string `json:"balance"`
	}

	QueryAccountRequest struct {
		AccountID uint64 `uri:"account_id" json:"account_id" binding:"required"`
	}

	// CreateAccountResponse represents the JSON response body structure
	QueryResponse struct {
		AccountID uint64 `json:"account_id"`
		Balance   string `json:"balance"`
	}
)
