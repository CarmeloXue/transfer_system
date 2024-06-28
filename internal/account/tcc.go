package account

import (
	"context"
	"errors"
	"fmt"
	"main/common/log"
	"main/common/utils"
	"main/model"
	. "main/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	ErrUnknowStage                   = errors.New("unknow fund movement status")
)

type TCC interface {
	Try(ctx context.Context, transactionID string, sourceAccountID, destinationAccountID int, amount int64) error

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
 * Try will make sure sender have enough balance to go out, and receiver have enough space to take this amount
 * */
func (s *tccService) Try(ctx context.Context, transactionID string, sourceAccountID, destinationAccountID int, amount int64) error {
	var (
		logger = log.GetSugger()
	)
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// check if transaction is already tried
		fundMovement, err := selectFundmovementForUpdate(tx, transactionID)
		// only proceed if no fund movement
		if err != nil && err == gorm.ErrRecordNotFound {

			sourceAcc, destAcc, err := loadAccounts(tx, sourceAccountID, destinationAccountID)
			if err != nil {
				return err
			}
			// lock source's amount
			err = sourceAcc.TryTransfer(tx, amount)
			if err != nil {
				if err == utils.ErrNegativeValue {
					err = ErrInsufficientBalance
				}
				return err
			}
			// lock reciever's income
			if err := destAcc.TryReceive(tx, amount); err != nil {
				return err
			}
			tried := FundMovement{
				TransactionID:        transactionID,
				SourceAccountID:      sourceAccountID,
				DestinationAccountID: destinationAccountID,
				Amount:               amount,
				Stage:                Tried,
			}
			// Create deduct fund movement.
			if err := tx.Model(FundMovement{}).Create(&tried).Error; err != nil {
				// race condition, other goroutine created it between last check and start transaction
				// just return normal nil to keep idempotent
				if err == gorm.ErrDuplicatedKey {
					return nil
				}
				return ErrFailedToWritePayment
			}

			logger.Info("try transaction success", "transactionID", transactionID, "amount", amount)
			return nil
		}

		if err != nil {
			return err
		}

		switch fundMovement.Stage {
		case Confirmed:
			return nil
		case Tried:
			return nil
		case Canceled:
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
		tried, err := selectFundmovementForUpdate(tx, transactionID)
		// call confirm before try is not allowed, so not check not found here
		if err != nil {
			logger.Error("failed to get fund movement status", "err", err)
			return err
		}
		switch tried.Stage {
		case Confirmed:
			return nil
		case Canceled:
			logger.Error("fund movement already rolledbacked")
			return ErrRollbacked
		case Tried:
			break
		default:
			return ErrUnknowStage
		}

		sourceAcc, destAcc, err := loadAccounts(tx, tried.SourceAccountID, tried.DestinationAccountID)
		if err != nil {
			return err
		}
		// confirm from source
		if err := sourceAcc.Transfer(tx, tried.Amount); err != nil {
			return err
		}

		// confirm from dest
		if err := destAcc.Recieve(tx, tried.Amount); err != nil {
			return err
		}

		// update func movement
		if err = tx.Model(FundMovement{}).Where("transaction_id = ?", tried.TransactionID).Update("stage", Confirmed).Error; err != nil {
			return ErrFMFailedToMoveDestConfirmed
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
		logger    = log.GetSugger()
		globalErr error
	)

	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tried, err := selectFundmovementForUpdate(tx, transactionID)

		// Cancel before try, put a rollback with 0 amount
		if err == gorm.ErrRecordNotFound {
			logger.Info("Cancel before try, save a rollback fm", "transactionID", transactionID)
			err = tx.Create(&FundMovement{
				TransactionID: transactionID,
				Stage:         Canceled,
			}).Error
			if err != nil {
				return err
			}
			globalErr = ErrEmptyRollback
			return nil
		}

		if err != nil {
			return err
		}

		switch tried.Stage {
		case Confirmed:
			return ErrConfirmed
		case Canceled:
			return nil
		case Tried:
			break
		default:
			return ErrUnknowStage
		}

		sourceAcc, destAcc, err := loadAccounts(tx, tried.SourceAccountID, tried.DestinationAccountID)
		if err != nil {
			return err
		}

		if err := sourceAcc.CancelTransfer(tx, tried.Amount); err != nil {
			return err
		}

		if err := destAcc.CancelRecieve(tx, tried.Amount); err != nil {
			return err
		}

		// Create refund fund movement
		if err := tx.Model(FundMovement{}).Where("transaction_id = ?", tried.TransactionID).Update("stage", Canceled).Error; err != nil {
			return ErrFailedToRollback
		}

		return nil
	})
	// Empty rollback
	if txErr == nil && globalErr != nil {
		return globalErr
	}
	return txErr
}

func selectFundmovementForUpdate(tx *gorm.DB, transactionID string) (*model.FundMovement, error) {
	var fundMovement FundMovement
	if err := tx.Model(FundMovement{}).Clauses(clause.Locking{Strength: "Update"}).First(&fundMovement, FundMovement{TransactionID: transactionID}).Error; err != nil {
		return nil, err
	}
	return &fundMovement, nil
}

func loadAccounts(tx *gorm.DB, sourceID, destID int) (*Account, *Account, error) {
	var (
		sourceAcc Account
		destAcc   Account
		err       error
	)
	if err = tx.Model(Account{}).Clauses(clause.Locking{Strength: "Update"}).First(&sourceAcc, Account{AccountID: sourceID}).Error; err != nil {
		return nil, nil, err
	}
	if err = tx.Model(Account{}).Clauses(clause.Locking{Strength: "Update"}).First(&destAcc, Account{AccountID: destID}).Error; err != nil {
		return nil, nil, err
	}

	return &sourceAcc, &destAcc, nil
}
