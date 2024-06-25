package model

import (
	"main/common/utils"
	"time"

	"gorm.io/gorm"
)

type FundMovementStage int32

var (
	Tried     FundMovementStage = 1
	Confirmed FundMovementStage = 2
	Canceled  FundMovementStage = 3
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
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID  int       `gorm:"unique;not null" json:"account_id"`
	Balance    int64     `gorm:"bigint;not null;default:0" json:"balance"`
	InBalance  int64     `gorm:"bigint;not null;default:0" json:"in_balance"`
	OutBalance int64     `gorm:"bigint;not null;default:0" json:"out_balance"`
	CreatedAt  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName sets the insert table name for this struct type.
func (Account) TableName() string {
	return "account_tab"
}

func (a *Account) TryTransfer(tx *gorm.DB, amount int64) error {
	// check if balance enough
	if _, err := utils.SafeAdd(a.Balance, -a.OutBalance, -amount); err != nil {
		return err
	}
	// return latest out balance
	outBalance, err := utils.SafeAdd(a.OutBalance, amount)
	if err != nil {
		return err
	}
	return tx.Model(Account{}).Where("account_id = ?", a.AccountID).Updates(map[string]interface{}{"out_balance": outBalance}).Error
}

func (a *Account) Transfer(tx *gorm.DB, amount int64) error {
	var (
		balance    int64
		outBalance int64
	)

	balance, err := utils.SafeAdd(a.Balance, -amount)
	if err != nil {
		return err
	}
	outBalance, err = utils.SafeAdd(a.OutBalance, -amount)
	if err != nil {
		return err
	}
	return tx.Model(Account{}).Where("account_id = ?", a.AccountID).Updates(map[string]interface{}{
		"balance":     balance,
		"out_balance": outBalance,
	}).Error
}

func (a *Account) TryReceive(tx *gorm.DB, amount int64) error {
	// check if exceed limit
	if _, err := utils.SafeAdd(a.Balance, a.InBalance, amount); err != nil {
		return err
	}
	// return latest in balance
	inBalance, err := utils.SafeAdd(a.InBalance, amount)
	if err != nil {
		return err
	}
	return tx.Model(Account{}).Where("account_id = ?", a.AccountID).Updates(map[string]interface{}{"in_balance": inBalance}).Error
}

func (a *Account) Recieve(tx *gorm.DB, amount int64) error {
	var (
		balance   int64
		inBalance int64
	)

	balance, err := utils.SafeAdd(a.Balance, amount)
	if err != nil {
		return err
	}
	inBalance, err = utils.SafeAdd(a.InBalance, -amount)
	if err != nil {
		return err
	}

	return tx.Model(Account{}).Where("account_id = ?", a.AccountID).Updates(map[string]interface{}{
		"balance":    balance,
		"in_balance": inBalance,
	}).Error
}

func (a *Account) CancelTransfer(tx *gorm.DB, amount int64) error {
	outBalance, err := utils.SafeAdd(a.OutBalance, -amount)
	if err != nil {
		return err
	}
	return tx.Model(Account{}).Where("account_id = ?", a.AccountID).Updates(map[string]interface{}{
		"out_balance": outBalance,
	}).Error
}

func (a *Account) CancelRecieve(tx *gorm.DB, amount int64) error {
	inBalance, err := utils.SafeAdd(a.InBalance, -amount)
	if err != nil {
		return err
	}
	return tx.Model(Account{}).Where("account_id = ?", a.AccountID).Updates(map[string]interface{}{
		"in_balance": inBalance,
	}).Error
}
