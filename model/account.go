package model

import "time"

type FundMovementStage int32

var (
	FMStageSourceOnHold FundMovementStage = 1
	FMStageDestConfirmd FundMovementStage = 2
	FMStageRollbacked   FundMovementStage = 3
)

type FundMovement struct {
	ID                   int               `gorm:"primaryKey;column:id" json:"id"`
	TransactionID        string            `gorm:"column:transaction_id" json:"transaction_id"`
	Stage                FundMovementStage `gorm:"column:stage" json:"stage"`
	SourceAccountID      int               `gorm:"column:source_account_id" json:"source_account_id"`
	DestinationAccountID int               `gorm:"column:destination_account_id" json:"destination_account_id"`
	Amount               int64             `gorm:"column:amount" json:"amount"`
	CreatedAt            time.Time         `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time         `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName sets the insert table name for this struct type.
func (FundMovement) TableName() string {
	return "fund_movement_tab"
}

type Account struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID int       `gorm:"unique;not null" json:"account_id"`
	Balance   int64     `gorm:"bigint;not null;default:0" json:"balance"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName sets the insert table name for this struct type.
func (Account) TableName() string {
	return "account_tab"
}
