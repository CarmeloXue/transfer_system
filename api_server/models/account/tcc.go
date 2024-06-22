package account

import (
	"context"
	"fmt"
	"main/common/log"
	"main/models/transaction"
	"math"
	"time"
)

const timeoutSeconds = 3

// Try will write a payment fund movement, then deduct from source user's amount
func (r *repository) Try(ctx context.Context, trx transaction.Transaction) error {
	txCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	tx := r.db.WithContext(txCtx).Begin()
	if tx.Error != nil {
		panic("failed to begin transaction")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var err error

	fm := FundMovement{
		TransactionID:        trx.ID,
		SourceAccountID:      trx.SourceAccountID,
		DestinationAccountID: trx.DestinationAccountID,
		Amount:               trx.Amount,
		Direction:            string(Payment),
	}
	// Create deduct fund movement
	if err = tx.Model(FundMovement{}).Create(&fm).Error; err != nil {
		tx.Rollback()
		return ErrFailedToWritePayment
	}

	// Deduct user's balance
	account := Account{}

	err = tx.First(&account, Account{
		AccountID: trx.SourceAccountID,
	}).Error
	if err = tx.Model(FundMovement{}).Create(&fm).Error; err != nil {
		tx.Rollback()
		return ErrFailedToLoadUser
	}

	if account.Balance < trx.Amount {
		tx.Rollback()
		return ErrInsufficientBalance
	}

	newBalance := account.Balance - trx.Amount
	err = tx.Model(Account{}).Where("account_id = ?", account.AccountID).Update("balance", newBalance).Error
	if account.Balance < trx.Amount {
		tx.Rollback()
		return ErrFailedToDeductSourceBalance
	}

	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		return ErrFailedToCommit
	}

	log.GetLogger().Info(fmt.Sprintf("try transaction success. Transaction:%v\n", trx))
	return nil
}

// Confirm will confirm a transaction by write a paymentreceived fund movement, and add amount to destination user
func (r *repository) Confirm(ctx context.Context, trx transaction.Transaction) error {
	txCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	tx := r.db.WithContext(txCtx).Begin()
	if tx.Error != nil {
		panic("failed to begin transaction")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var err error

	fm := FundMovement{
		TransactionID:        trx.ID,
		SourceAccountID:      trx.SourceAccountID,
		DestinationAccountID: trx.DestinationAccountID,
		Amount:               trx.Amount,
		Direction:            string(PaymentReceived),
	}
	// Create deduct fund movement
	if err = tx.Model(FundMovement{}).Create(&fm).Error; err != nil {
		tx.Rollback()
		return ErrFailedToWritePayment
	}

	// Deduct user's balance
	account := Account{}

	err = tx.First(&account, Account{
		AccountID: trx.DestinationAccountID,
	}).Error
	if err = tx.Model(FundMovement{}).Create(&fm).Error; err != nil {
		tx.Rollback()
		return ErrFailedToLoadUser
	}

	if account.Balance+trx.Amount > math.MaxFloat64 {
		tx.Rollback()
		return ErrInsufficientBalance
	}

	newBalance := account.Balance + trx.Amount
	if err = tx.Model(Account{}).Where("account_id = ?", account.AccountID).Update("balance", newBalance).Error; err != nil {
		tx.Rollback()
		return ErrFailedToAddDestinationBalance
	}

	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		return ErrFailedToCommit
	}

	log.GetLogger().Info(fmt.Sprintf("confirm transaction success. Transaction:%v\n", trx))
	return nil
}

func (r *repository) Cancel(ctx context.Context, trx transaction.Transaction) error {

}
