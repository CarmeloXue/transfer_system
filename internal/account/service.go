package account

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"main/tools/currency"
	"main/tools/log"
	"sync"

	. "main/internal/model/account"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Service interface {
	CreateAccount(ctx context.Context, req CreateAccountRequest) error
	QueryAccount(ctx context.Context, req QueryAccountRequest) (Account, error)
}

type accountService struct {
	repo  AccountRepository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) Service {
	return &accountService{
		repo:  NewRepository(db),
		cache: cache,
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
	bs, _ := s.cache.Get(ctx, s.getAccountIdCacheKey(int(req.AccountID))).Bytes()

	if len(bs) != 0 {
		var account Account
		if err := json.NewDecoder(bytes.NewReader(bs)).Decode(&account); err == nil {
			return account, nil
		}
	}

	// if read from redis have issue, read from db
	return s.repo.GetAccountByID(ctx, int(req.AccountID))
}

func (s *accountService) getAccountIdCacheKey(accountID int) string {
	return fmt.Sprintf("cache_key-account-%d", accountID)
}
