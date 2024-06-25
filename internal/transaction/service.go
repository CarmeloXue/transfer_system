package transaction

import (
	"context"
	"errors"
	"main/common/log"
	"main/common/recovery"
	"main/common/utils"
	"main/internal/account"
	"main/model"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	DefaultCreateTransactionTimeoutSeconds = 3
	DefaultTryTransactionTimeoutSeconds    = 1
	MaxRetry                               = 3
)

var (
	ErrSameAccountTransactions = errors.New("source and destination cannot be the same")
)

type Service interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) (model.Transaction, error)
	QueryTransaction(ctx context.Context, req QueryTransactionRequest) (model.Transaction, error)
	RetryTransaction(ctx context.Context, req QueryTransactionRequest) (model.Transaction, error)

	// ConfirmTransaction(req ConfirmTransactionRequest) error
}

type service struct {
	repo        Repository
	accountTCC  account.TCC
	accountRepo account.AccountRepository
}

func NewService(repo Repository, accountTCC account.TCC, accountRepo account.AccountRepository) Service {
	return &service{repo: repo, accountTCC: accountTCC, accountRepo: accountRepo}
}

func (s *service) CreateTransaction(ctx context.Context, req CreateTransactionRequest) (model.Transaction, error) {
	if req.DestinationAccountID == req.SourceAccountID {
		return model.Transaction{}, ErrSameAccountTransactions
	}
	inflatedValue, err := utils.ParseString(req.Amount)
	if err != nil {
		return model.Transaction{}, err
	}

	if _, err := s.accountRepo.GetAccountByID(ctx, req.SourceAccountID); err != nil {
		return model.Transaction{}, errors.New("invalid sender")
	}

	if _, err := s.accountRepo.GetAccountByID(ctx, req.DestinationAccountID); err != nil {
		return model.Transaction{}, errors.New("invalid reciever")
	}

	trx := model.Transaction{
		SourceAccountID:      req.SourceAccountID,
		DestinationAccountID: req.DestinationAccountID,
		Amount:               inflatedValue,
		TransactionID:        utils.GenerateTransactionID(),
		TransactionStatus:    model.Pending,
	}
	timeoutSeconds := viper.GetInt("create_transaction_timeout")
	tCtx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeoutSeconds))
	defer cancel()
	// Create pending transaction
	err = s.repo.CreateTransaction(tCtx, trx)
	if err != nil {
		return model.Transaction{}, err
	}

	trxChan, err := s.processTransaction(tCtx, &trx)

	select {
	case tx, ok := <-trxChan:
		if ok {
			return tx, err
		}
	case <-tCtx.Done():
		return trx, err
	}
	return trx, err
}

func (s *service) QueryTransaction(ctx context.Context, req QueryTransactionRequest) (model.Transaction, error) {
	return s.repo.GetTransactionByID(ctx, req.TransactionID)
}

// TODO: Better expose this as a cmd, not a http request
func (s *service) RetryTransaction(ctx context.Context, req QueryTransactionRequest) (model.Transaction, error) {
	tx, err := s.repo.GetTransactionByID(ctx, req.TransactionID)
	if err != nil {
		return model.Transaction{}, err
	}
	tCtx, cancel := context.WithTimeout(ctx, time.Second*DefaultCreateTransactionTimeoutSeconds)
	defer cancel()
	// Start from try
	if tx.TransactionStatus == model.Pending || tx.TransactionStatus == model.Processing {
		trxChan, err := s.processTransaction(tCtx, &tx)

		select {
		case tx, ok := <-trxChan:
			if ok {
				return tx, err
			}
		case <-tCtx.Done():
			return tx, nil
		}
		return tx, nil
	}

	return tx, nil
}

func (s *service) processTransaction(ctx context.Context, transaction *model.Transaction) (<-chan model.Transaction, error) {
	err := s.tryWithTimeout(ctx, transaction)
	transactionChan := make(chan model.Transaction)
	go func() {
		defer close(transactionChan)
		defer func() {
			tx, err := s.repo.GetTransactionByID(ctx, transaction.TransactionID)
			if err == nil {
				transactionChan <- tx
			}
		}()
		defer recovery.GoRecovery()

		transaction.Retries = viper.GetInt("max_retries")
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				s.retryCancel(ctx, transaction)
			} else {
				_ = s.repo.UpdateTransactionStatus(ctx, transaction.TransactionID, model.Failed)
			}
			return
		}

		_ = s.repo.UpdateTransactionStatus(ctx, transaction.TransactionID, model.Processing)

		s.retryConfirm(ctx, transaction)
	}()

	return transactionChan, err
}

func (s *service) retryCancel(ctx context.Context, tx *model.Transaction) {
	log.GetSugger().Info("start to cancel transaction ", "transaction", tx)
	var err error
	for i := 0; i < tx.Retries; i++ {
		log.GetSugger().Info("call cancel ", "transaction", tx.TransactionID, "err", err)
		err = s.accountTCC.Cancel(ctx, tx.TransactionID)
		log.GetSugger().Info("try cancel ", "transaction", tx.TransactionID, "err", err)
		if err == nil || err == account.ErrEmptyRollback {
			if err = s.repo.UpdateTransactionStatus(ctx, tx.TransactionID, model.Failed); err == nil {
				return
			}
		}
		time.Sleep(30 * time.Millisecond)
	}
	if err != nil {
		log.GetSugger().Error("failed to cancel transaction ", "transaction", tx, "err", err)
	}

}

func (s *service) retryConfirm(ctx context.Context, tx *model.Transaction) {
	log.GetLogger().With(zap.Any("transaction", tx)).Info("prepare to confirm")
	var err error
	for i := 0; i < tx.Retries; i++ {
		err = s.accountTCC.Confirm(ctx, tx.TransactionID)
		if err == nil {
			if err = s.repo.UpdateTransactionStatus(ctx, tx.TransactionID, model.Fulfiled); err == nil {
				return
			}
		}
		time.Sleep(30 * time.Millisecond)
	}
	if err != nil {
		log.GetSugger().Error("failed to confirm transaction ", "transaction", tx, "err", err)
	}
}

func (s *service) try(ctx context.Context, tx *model.Transaction) <-chan error {
	errChan := make(chan error)

	go func() {
		defer close(errChan)
		if err := s.accountTCC.Try(ctx, tx.TransactionID, tx.SourceAccountID, tx.DestinationAccountID, tx.Amount); err != nil {
			errChan <- err
		}
	}()

	return errChan
}

func (s *service) tryWithTimeout(ctx context.Context, tx *model.Transaction) error {
	timeOutCtx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(viper.GetInt("try_timeout")))
	defer cancel()

	errChan := s.try(ctx, tx)

	select {
	case <-timeOutCtx.Done():
		return timeOutCtx.Err()
	case err := <-errChan:
		return err
	}
}
