package transaction

import "main/internal/account"

type Service interface {
	CreateTransaction(req CreateTransactionRequest) error
	ConfirmTransaction(req ConfirmTransactionRequest) error
}

type service struct {
	repo       Repository
	accountTCC account.TCC
}
