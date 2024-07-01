package account

import (
	"context"
	"main/common/log"
	"main/tools/currency"
	"sync"

	. "main/model"

	"gorm.io/gorm"
)

type Service interface {
	CreateAccount(ctx context.Context, req CreateAccountRequest) error
	QueryAccount(ctx context.Context, req QueryAccountRequest) (Account, error)
}

type accountService struct {
	repo AccountRepository
}

func NewService(db *gorm.DB) Service {
	return &accountService{
		repo: NewRepository(db),
	}
}

var (
	acService *accountService
	once      sync.Once
)

func NewAccountService(db *gorm.DB) *accountService {
	once.Do(func() {
		acService = &accountService{repo: NewRepository(db)}
	})
	return acService
}

func (s *accountService) CreateAccount(ctx context.Context, req CreateAccountRequest) error {
	inflatedValue, err := currency.ParseString(req.InitialBalance)
	if err != nil {
		log.GetLogger().Error(err.Error())
		return err
	}

	err = s.repo.CreateAccount(ctx, &Account{
		AccountID: int(req.AccountID),
		Balance:   inflatedValue,
	})
	if err != nil {
		log.GetLogger().Error(err.Error())
		return err
	}

	return nil
}

func (s *accountService) QueryAccount(ctx context.Context, req QueryAccountRequest) (Account, error) {
	return s.repo.GetAccountByID(ctx, int(req.AccountID))
}
