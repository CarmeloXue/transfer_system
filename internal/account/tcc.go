package account

import (
	"context"
	"errors"
	"fmt"
	"main/common/log"
	"time"

	"gorm.io/gorm"
)

const timeoutSeconds = 3

var (
	ErrFailedToWritePayment          = errors.New("failed to write payment")
	ErrFailedToLoadUser              = errors.New("failed to load user")
	ErrInsufficientBalance           = errors.New("insufficient balance")
	ErrFailedToDeductSourceBalance   = errors.New("failed to deduct source balance")
	ErrFailedToCommit                = errors.New("failed to commit")
	ErrExceedingMaxAmount            = errors.New("exceeding max amount")
	ErrFailedToAddDestinationBalance = errors.New("failed to add destination balance")
	ErrTransactionTried              = errors.New("transaction already tried")
	ErrInternalError                 = errors.New("internal error")
)

type TCC interface {
	// TCC api
	Try(ctx context.Context, transactionID, sourceAccountID, destinationAccountID int, amount float64) error
}

type tccService struct {
	db *gorm.DB
}

func NewTCCService(db *gorm.DB) TCC {
	return &tccService{db: db}
}

// Try will write a payment fund movement, then deduct from source user's amount
func (r *tccService) Try(ctx context.Context, transactionID, sourceAccountID, destinationAccountID int, amount float64) error {
	txCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	// check if transaction is already tried
	repo := NewRepository(r.db)
	if _, err := repo.GetFundMovement(ctx, transactionID, sourceAccountID, Payment); err == nil {
		return ErrTransactionTried
	} else {
		if err != gorm.ErrRecordNotFound {
			return ErrInternalError
		}
	}

	tx := r.db.WithContext(txCtx).Begin()
	if tx.Error != nil {
		panic("failed to begin transaction")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	payment := FundMovement{
		TransactionID:        transactionID,
		SourceAccountID:      sourceAccountID,
		DestinationAccountID: destinationAccountID,
		Amount:               amount,
		FundMovementType:     string(Payment),
	}
	// Create deduct fund movement
	if err := tx.Model(FundMovement{}).Create(&payment).Error; err != nil {
		tx.Rollback()
		return ErrFailedToWritePayment
	}

	// Deduct user's balance
	account := Account{}
	if err := tx.First(&account, Account{
		AccountID: sourceAccountID,
	}).Error; err != nil {
		tx.Rollback()
		return ErrFailedToLoadUser
	}

	if account.Balance < amount {
		tx.Rollback()
		return ErrInsufficientBalance
	}

	newBalance := account.Balance - amount
	if err := tx.Model(Account{}).Where("account_id = ?", account.AccountID).Update("balance", newBalance).Error; err != nil {
		tx.Rollback()
		return ErrFailedToDeductSourceBalance
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return ErrFailedToCommit
	}

	log.GetLogger().Info(fmt.Sprintf("try transaction success. transaction: %v amount: %v\n", transactionID, amount))
	return nil
}

func (s *tccService) Confirm(ctx context.Context, transactionID string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Implement the Confirm logic (finalize the transfer, etc.)
		return nil
	})
}

func (s *tccService) Cancel(ctx context.Context, transactionID string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Implement the Cancel logic (rollback the reservation, etc.)
		return nil
	})
}
