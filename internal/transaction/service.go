package transaction

import (
	"context"
	"errors"
	"fmt"
	"main/common/log"
	"main/common/utils"
	"main/internal/account"
	"main/model"
	"time"

	"go.uber.org/zap"
)

const (
	DefaultCreateTransactionTimeoutSeconds = 3
	MaxRetry                               = 5
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
	repo       Repository
	accountTCC account.TCC
}

func NewService(repo Repository, accountTCC account.TCC) Service {
	return &service{repo: repo, accountTCC: accountTCC}
}

func (s *service) CreateTransaction(ctx context.Context, req CreateTransactionRequest) (model.Transaction, error) {
	if req.DestinationAccountID == req.SourceAccountID {
		return model.Transaction{}, ErrSameAccountTransactions
	}
	float64Value, err := utils.ParseFloat64String(req.Amount)
	if err != nil {
		return model.Transaction{}, err
	}

	trx := model.Transaction{
		SourceAccountID:      req.SourceAccountID,
		DestinationAccountID: req.DestinationAccountID,
		Amount:               float64Value,
		TransactionID:        utils.GenerateTransactionID(),
		Status:               string(model.Pending),
	}
	tCtx, cancel := context.WithTimeout(ctx, time.Second*DefaultCreateTransactionTimeoutSeconds)
	defer cancel()
	err = s.repo.CreateTransaction(tCtx, trx)
	if err != nil {
		return model.Transaction{}, err
	}

	trxChan, err := s.processTransaction(tCtx, trx.TransactionID)
	if err != nil {
		return model.Transaction{}, err
	}
	select {
	case tx, ok := <-trxChan:
		if ok {
			return tx, nil
		}
	case <-tCtx.Done():
		return trx, nil
	}
	return trx, nil

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
	// Start from try
	if tx.Status == string(model.Pending) || tx.Status == string(model.Processing) {
		go s.processTransaction(ctx, tx.TransactionID)
	}

	time.Sleep(time.Second * 1)
	return s.repo.GetTransactionByID(ctx, req.TransactionID)
}

func (s *service) processTransaction(ctx context.Context, transactionID string) (<-chan model.Transaction, error) {

	tx, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("Failed to find transaction: %v", err))
		return nil, err
	}
	transactionChan := make(chan model.Transaction)

	go func() {
		defer close(transactionChan)
		defer func() {
			tx, err := s.repo.GetTransactionByID(ctx, transactionID)
			if err == nil {
				transactionChan <- tx
			}
		}()
		err = s.accountTCC.Try(ctx, tx.TransactionID, tx.SourceAccountID, tx.DestinationAccountID, tx.Amount)
		tx.Retries = MaxRetry
		log.GetLogger().Info(fmt.Sprintf("Try, err %v", err))
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				s.retryCancel(ctx, &tx)
			} else {
				s.updateStatus(ctx, tx.TransactionID, model.Failed)
			}
			return
		}
		if tx.Status == string(model.Pending) {
			s.updateStatus(ctx, tx.TransactionID, model.Processing)
		}
		s.retryConfirm(ctx, &tx)
	}()
	return transactionChan, err
}

func (s *service) retryCancel(ctx context.Context, tx *model.Transaction) {
	for i := 0; i < tx.Retries; i++ {
		err := s.accountTCC.Cancel(ctx, tx.TransactionID)
		if err == nil {
			s.updateStatus(ctx, tx.TransactionID, model.Failed)
			return
		}
		time.Sleep(1 * time.Second)
	}
	// TODO: alert
}

func (s *service) retryConfirm(ctx context.Context, tx *model.Transaction) {
	log.GetLogger().With(zap.Any("transaction", tx)).Info("prepare to confirm")
	for i := 0; i < tx.Retries; i++ {
		err := s.accountTCC.Confirm(ctx, tx.TransactionID)
		if err == nil {
			s.updateStatus(ctx, tx.TransactionID, model.Fulfiled)
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func (s *service) updateStatus(ctx context.Context, transactionID string, status model.TransactionStatus) {
	s.repo.UpdateTransactionStatus(ctx, transactionID, string(status))
}
