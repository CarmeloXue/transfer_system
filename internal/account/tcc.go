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
	ErrFailedToWritePaymentReceived  = errors.New("failed to write payment received")
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
	logger := log.GetSugger()
	// check if transaction is already tried
	repo := NewRepository(s.db)
	if _, err := repo.GetFundMovement(ctx, FundMovement{
		TransactionID:    transactionID,
		FundMovementType: string(FMPayment),
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
		payment := FundMovement{
			TransactionID:        transactionID,
			SourceAccountID:      sourceAccountID,
			DestinationAccountID: destinationAccountID,
			Amount:               amount,
			FundMovementType:     string(FMPayment),
		}
		// Create deduct fund movement
		if err := tx.Model(FundMovement{}).Create(&payment).Error; err != nil {
			logger.Error("failed to create payment FM", "err", err)
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
	validator, err := s.getFundMovementValidator(ctx, transactionID)
	if err != nil {
		logger.Error("failed to get fm validator", "err", err)
		return err
	}
	isFinal, err := validator.isTransactionFinal()
	// TODO: Need to send alert to trigger manual check
	if err != nil {
		logger.Error("suspicious fund movemnet transaction", "err", err, "transaction", transactionID)
		return err
	}

	if isFinal {
		return nil
	}
	// check if transaction is already tried
	log.GetLogger().Info(fmt.Sprintf("start confirm transaction %v", transactionID))
	repo := NewRepository(s.db)
	payment := validator.getPayment()

	if _, err = repo.GetFundMovement(ctx, FundMovement{
		TransactionID:    transactionID,
		FundMovementType: string(FMPaymentReceived),
	}); err == nil {
		// if already confirmed, don't need to do anything
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return ErrInternalError
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		paymentRecieved := &FundMovement{
			TransactionID:        payment.TransactionID,
			SourceAccountID:      payment.SourceAccountID,
			DestinationAccountID: payment.DestinationAccountID,
			Amount:               payment.Amount,
			FundMovementType:     string(FMPaymentReceived),
		}

		if err = tx.Model(FundMovement{}).Create(paymentRecieved).Error; err != nil {
			return ErrFailedToWritePaymentReceived
		}

		if err = updateAccountBalance(tx, paymentRecieved.DestinationAccountID, payment.Amount); err != nil {
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
	validator, err := s.getFundMovementValidator(ctx, transactionID)
	if err != nil {
		return err
	}
	isFinal, err := validator.isTransactionFinal()
	// TODO: Need to send alert to trigger manual check
	if err != nil {
		return err
	}

	// only transaction in middle way need to cancel
	if isFinal || validator.isTransactionPending() {
		return nil
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		payment := validator.getPayment()
		paymentRefund := &FundMovement{
			TransactionID:        payment.TransactionID,
			SourceAccountID:      payment.SourceAccountID,
			DestinationAccountID: payment.DestinationAccountID,
			Amount:               payment.Amount,
			FundMovementType:     string(FMPaymentRefund),
		}
		// Create refund fund movement
		if err := tx.Model(FundMovement{}).Create(paymentRefund).Error; err != nil {
			return ErrFailedToWritePayment
		}
		// add back amount to source account
		if err := updateAccountBalance(tx, paymentRefund.SourceAccountID, payment.Amount); err != nil {
			return err
		}

		return nil
	})
}

func (s *tccService) getFundMovementValidator(ctx context.Context, transactionID string) (*fundMovementValidator, error) {
	fundMvmts, err := NewRepository(s.db).QueryFundMovement(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	return newFundMovementValidator(fundMvmts)
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

type fundMovementValidator struct {
	inFlow        int
	outFlow       int
	payment       *FundMovement
	transactionID string
}

func newFundMovementValidator(fms []FundMovement) (*fundMovementValidator, error) {
	validator := &fundMovementValidator{}

	for idx, fm := range fms {
		switch fm.FundMovementType {
		case string(FMPayment):
			validator.inFlow++
			validator.payment = &fms[idx]
			validator.transactionID = fm.TransactionID
		case string(FMPaymentRefund), string(FMPaymentReceived):
			validator.outFlow++
			validator.transactionID = fm.TransactionID
		default:
			return nil, errors.New(fmt.Sprintf("unsupported movement type %v", fm.FundMovementType))
		}
	}
	return validator, nil
}

/**
 * Checks if current transaction is cancellable, according to fund movement records.
 */
func (v *fundMovementValidator) isTransactionFinal() (bool, error) {
	if v.inFlow > 1 || v.outFlow > 1 || (v.inFlow == 0 && v.outFlow != 0) {
		return false, errors.New(fmt.Sprintf("fatal errors, fund movements not correct under transaction %v", v.transactionID))
	}

	if (v.inFlow == v.outFlow) && v.inFlow != 0 {
		return true, nil
	}

	return false, nil
}

func (v *fundMovementValidator) isTransactionPending() bool {
	if v.inFlow == 0 && v.outFlow == 0 {
		return true
	}
	return false
}

func (v *fundMovementValidator) getPayment() *FundMovement {
	return v.payment
}
