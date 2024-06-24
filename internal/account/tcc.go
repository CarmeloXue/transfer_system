package account

import (
	"context"
	"errors"
	"fmt"
	"main/common/log"
	. "main/model"

	"gorm.io/gorm"
)

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
	ErrEmptyRollback                 = errors.New("empty rollback")
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

/**
 * Try will write a payment fund movement, then deduct from source user's amount
 *
 * nil error indicates fund movement is SourceOnHold, and source account is deducted.
 * ErrRollbacked indicates there already an emtpy rollback or real rollback.
 * */
func (s *tccService) Try(ctx context.Context, transactionID string, sourceAccountID, destinationAccountID int, amount float64) error {
	var (
		logger = log.GetSugger()
	)
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// check if transaction is already tried
		var fundMovement FundMovement
		err := tx.Model(FundMovement{}).First(&fundMovement, FundMovement{TransactionID: transactionID}).Error
		// only proceed if no fund movement
		if err != nil && err == gorm.ErrRecordNotFound {
			sourceOnHold := FundMovement{
				TransactionID:        transactionID,
				SourceAccountID:      sourceAccountID,
				DestinationAccountID: destinationAccountID,
				Amount:               amount,
				Stage:                FMStageSourceOnHold,
			}
			// Create deduct fund movement.
			if err := tx.Model(FundMovement{}).Create(&sourceOnHold).Error; err != nil {
				// race condition, other goroutine created it between last check and start transaction
				// just return normal nil to keep idempotent
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
		}

		if err != nil {
			return err
		}

		switch fundMovement.Stage {
		case FMStageDestConfirmd:
			return nil
		case FMStageSourceOnHold:
			return nil
		case FMStageRollbacked:
			return ErrRollbacked
		default:
			return errors.New("unknow fund movement status")
		}
	})
}

/**
 * Confirm used to send funds to destination accounts after check fundmovement
 * If had payment_recieved, will just return success to keep idenpotent.
 * If confirm failed, can retry.
 *
 * nil return means confirm fund movement is destConfirmed and fund added to destination account
 * ErrRollbacked indicate try to confirm a canceled transaction.
 */
func (s *tccService) Confirm(ctx context.Context, transactionID string) error {
	var (
		logger = log.GetSugger()
	)
	// check if transaction is already tried
	log.GetLogger().Info(fmt.Sprintf("start confirm transaction %v", transactionID))
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var destConfirmed FundMovement
		err := tx.Model(FundMovement{}).First(&destConfirmed, FundMovement{TransactionID: transactionID}).Error

		// call confirm before try is not allowed, so not check not found here
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
 * If cancel before try, return a empty cancel error.
 * If the fundmovements record is not valie, which means it will be a big bug in this system, need to alert admin to check
 * If the fundmovements looks fine, create a refund fundmovement and add amount in payment back to source acount
 *
 *  ErrConfirmed indicates transaction confirmed. Don't need to modify transaction
 *  nil indicates fund movement is rollbacked, adn fund is returned to source. Upstream can change transaction status to refund
 *
 */
func (s *tccService) Cancel(ctx context.Context, transactionID string) error {
	var (
		logger = log.GetSugger()
		err    error
	)

	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rollback FundMovement
		err = tx.Model(FundMovement{}).First(&rollback, FundMovement{TransactionID: transactionID}).Error

		// Cancel before try, put a rollback with 0 amount
		if err == gorm.ErrRecordNotFound {
			logger.Info("Cancel before try, save a rollback fm", "transactionID", transactionID)
			err = tx.Create(&FundMovement{
				TransactionID: transactionID,
				Stage:         FMStageRollbacked,
			}).Error
			if err != nil {
				return err
			}
			err = ErrEmptyRollback
			return nil
		}

		if err != nil {
			return err
		}

		if rollback.Stage == FMStageDestConfirmd {
			return ErrConfirmed
		}

		if rollback.Stage == FMStageRollbacked {
			return nil
		}

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
	// Empty rollback
	if txErr == nil && err != nil {
		return err
	}
	return txErr
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
