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
	ErrPaymentNotDone                = errors.New("payment not done")
	ErrFailedToWritePaymentReceived  = errors.New("failed to write payment received")
)

type TCC interface {
	Try(ctx context.Context, transactionID, sourceAccountID, destinationAccountID int, amount float64) error

	Confirm(ctx context.Context, transactionID int) error

	Cancel(ctx context.Context, transactionID int) error
}

type tccService struct {
	db *gorm.DB
}

func NewTCCService(db *gorm.DB) TCC {
	return &tccService{db: db}
}

// Try will write a payment fund movement, then deduct from source user's amount
func (s *tccService) Try(ctx context.Context, transactionID, sourceAccountID, destinationAccountID int, amount float64) error {
	// check if transaction is already tried
	repo := NewRepository(s.db)
	if _, err := repo.GetFundMovement(ctx, FundMovement{
		TransactionID:    transactionID,
		FundMovementType: string(Payment),
	}); err == nil {
		return ErrTransactionTried
	} else {
		if err != gorm.ErrRecordNotFound {
			return ErrInternalError
		}
	}

	txCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return s.db.WithContext(txCtx).Transaction(func(tx *gorm.DB) error {
		payment := FundMovement{
			TransactionID:        transactionID,
			SourceAccountID:      sourceAccountID,
			DestinationAccountID: destinationAccountID,
			Amount:               amount,
			FundMovementType:     string(Payment),
		}
		// Create deduct fund movement
		if err := tx.Model(FundMovement{}).Create(&payment).Error; err != nil {
			return ErrFailedToWritePayment
		}

		if err := updateAccountBalance(tx, sourceAccountID, -amount); err != nil {
			return err
		}

		log.GetLogger().Info(fmt.Sprintf("try transaction success. transaction: %v amount: %v\n", transactionID, amount))
		return nil
	})
}

/**
 * Confirm used to send funds to destination accounts after check fundmovement
 * If had payment_recieved, will just return success to keep idenpotent. 
 * If confirm failed, can retry.
 */
func (s *tccService) Confirm(ctx context.Context, transactionID int) error {
	var (
		err     error
		payment FundMovement
	)
	// check if transaction is already tried
	repo := NewRepository(s.db)
	if payment, err = repo.GetFundMovement(ctx, FundMovement{
		TransactionID:    transactionID,
		FundMovementType: string(Payment),
	}); err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrPaymentNotDone
		}
		return ErrInternalError
	}

	if _, err = repo.GetFundMovement(ctx, FundMovement{
		TransactionID:    transactionID,
		FundMovementType: string(PaymentReceived),
	}); err == nil {
		// if already confirmed, don't need to do anything
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return ErrInternalError
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		paymentRecieved := payment
		paymentRecieved.FundMovementType = string(PaymentReceived)

		if err = tx.Model(FundMovement{}).Create(&paymentRecieved).Error; err != nil {
			return ErrFailedToWritePaymentReceived
		}

		if err = updateAccountBalance(tx, paymentRecieved.DestinationAccountID, paymentRecieved.Amount); err != nil {
			return err
		}

		return nil
	})
}

/**
 * Cancale a transaction
 * If cancel before try, just return success.
 * If the fundmovements record is not valie, which means it will be a big bug in this system, need to alert admin to check
 * If the fundmovements looks fine, create a refund fundmovement and add amount in payment back to source acount
 */
func (s *tccService) Cancel(ctx context.Context, transactionID int) error {
	fundMvmts, err := NewRepository(s.db).QueryFundMovement(ctx, transactionID)
	if err != nil {
		return err
	}
	cancelValidators := cancelValidator(fundMvmts)
	needCancel, err := cancelValidators.needToCancel()

	// TODO: Need to send alert to trigger manual check
	if err != nil {
		return err
	}

	if !needCancel {
		return nil
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		payment := cancelValidators.getPayment()
		paymentRefund := payment
		paymentRefund.FundMovementType = string(PaymentRefund)
		// Create refund fund movement
		if err := tx.Model(FundMovement{}).Create(&paymentRefund).Error; err != nil {
			return ErrFailedToWritePayment
		}
		// add back amount to source account
		if err := updateAccountBalance(tx, paymentRefund.SourceAccountID, payment.Amount); err != nil {
			return err
		}

		return nil
	})
}

func updateAccountBalance(tx *gorm.DB, accountID int, amount float64) error {
	// Deduct user's balance
	account := Account{}
	if err := tx.First(&account, Account{
		AccountID: accountID,
	}).Error; err != nil {
		return ErrFailedToLoadUser
	}
	newBalance := account.Balance + amount
	if newBalance < 0 {
		return ErrInsufficientBalance
	}

	if err := tx.Model(Account{}).Where("account_id = ?", account.AccountID).Update("balance", newBalance).Error; err != nil {
		return ErrFailedToDeductSourceBalance
	}
	return nil
}

type cancelValidator []FundMovement

func (v cancelValidator) needToCancel() (bool, error) {

	return true, nil
}

func (v cancelValidator) getPayment() FundMovement {

	return FundMovement{}
}
