package account

import (
	"context"
	"errors"
	"fmt"
	"main/common/log"
	. "main/model"
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
	ErrFMFailedToMoveDestConfirmed   = errors.New("failed to move confirmed")
	ErrRollbacked                    = errors.New("rollbacked")
	ErrConfirmed                     = errors.New("confirmed")
	ErrFailedToRollback              = errors.New("failed to rollback")
)

type TCC interface {
	Try(ctx context.Context, transactionID string, sourceAccountID, destinationAccountID int, amount float64) error

	Confirm(ctx context.Context, transactionID string) error

	Cancel(ctx context.Context, transactionID string) error
}

type tccService struct {
	db *gorm.DB
}

func NewTCCService(db *gorm.DB) TCC {
	return &tccService{db: db}
}

// Try will write a payment fund movement, then deduct from source user's amount
func (s *tccService) Try(ctx context.Context, transactionID string, sourceAccountID, destinationAccountID int, amount float64) error {
	var (
		logger = log.GetSugger()
		// check if transaction is already tried
		repo = NewRepository(s.db)
	)

	// TODO use goroutine. Currently met a issue in mock db connection in test
	if _, err := repo.GetAccountByID(ctx, sourceAccountID); err != nil {
		return errors.New("invalid sender")
	}

	if _, err := repo.GetAccountByID(ctx, destinationAccountID); err != nil {
		return errors.New("invalid reciever")
	}

	if _, err := repo.GetFundMovement(ctx, FundMovement{
		TransactionID: transactionID,
	}); err == nil {
		return nil
	} else {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	}

	txCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return s.db.WithContext(txCtx).Transaction(func(tx *gorm.DB) error {
		sourceOnHold := FundMovement{
			TransactionID:        transactionID,
			SourceAccountID:      sourceAccountID,
			DestinationAccountID: destinationAccountID,
			Amount:               amount,
			Stage:                FMStageSourceOnHold,
		}
		// Create deduct fund movement
		if err := tx.Model(FundMovement{}).Create(&sourceOnHold).Error; err != nil {
			if err == gorm.ErrDuplicatedKey {
				return nil
			}
			return ErrFailedToWritePayment
		}

		if err := updateAccountBalance(tx, sourceAccountID, -amount); err != nil {
			logger.Error("failed to update balance", "err", err)
			return err
		}

		logger.Info("try transaction success", "transactionID", transactionID, "amount", amount)
		return nil
	})
}

/**
 * Confirm used to send funds to destination accounts after check fundmovement
 * If had payment_recieved, will just return success to keep idenpotent.
 * If confirm failed, can retry.
 */
func (s *tccService) Confirm(ctx context.Context, transactionID string) error {
	var (
		err    error
		logger = log.GetSugger()
	)
	repo := NewRepository(s.db)
	destConfirmed, err := repo.GetFundMovement(ctx, FundMovement{
		TransactionID: transactionID,
	})
	if err != nil {
		logger.Error("failed to get fund movement status", "err", err)
		return err
	}

	if destConfirmed.Stage == FMStageDestConfirmd {
		return nil
	}

	if destConfirmed.Stage == FMStageRollbacked {
		logger.Error("fund movement already rolledbacked")
		return ErrRollbacked
	}
	// check if transaction is already tried
	log.GetLogger().Info(fmt.Sprintf("start confirm transaction %v", transactionID))

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err = tx.Model(FundMovement{}).Where("transaction_id = ?", destConfirmed.TransactionID).Update("stage", FMStageDestConfirmd).Error; err != nil {
			return ErrFMFailedToMoveDestConfirmed
		}

		if err = updateAccountBalance(tx, destConfirmed.DestinationAccountID, destConfirmed.Amount); err != nil {
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
func (s *tccService) Cancel(ctx context.Context, transactionID string) error {
	var (
		err    error
		logger = log.GetSugger()
	)
	repo := NewRepository(s.db)
	rollback, err := repo.GetFundMovement(ctx, FundMovement{
		TransactionID: transactionID,
	})
	if err != nil && err != gorm.ErrRecordNotFound {

		return err
	}

	// Cancel before try, put a rollback with 0 amount
	if err == gorm.ErrRecordNotFound {
		logger.Info("Cancel before try, save a rollback fm", "transactionID", transactionID)
		return repo.CreateFundMovement(ctx, &FundMovement{
			TransactionID: transactionID,
			Stage:         FMStageRollbacked,
		})
	}

	if rollback.Stage == FMStageDestConfirmd {
		return ErrConfirmed
	}

	if rollback.Stage == FMStageRollbacked {
		return nil
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// Create refund fund movement
		if err := tx.Model(FundMovement{}).Where("transaction_id = ?", rollback.TransactionID).Update("stage", FMStageRollbacked).Error; err != nil {
			return ErrFailedToRollback
		}
		// add back amount to source account
		if err := updateAccountBalance(tx, rollback.SourceAccountID, rollback.Amount); err != nil {
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
