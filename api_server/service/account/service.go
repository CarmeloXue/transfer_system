package account

import (
	"main/models/account"
	"sync"

	"gorm.io/gorm"
)

type accountService struct {
	repo account.AccountRepository
}

var (
	acService *accountService
	once      sync.Once
)

func NewAccountService(db *gorm.DB) *accountService {
	once.Do(func() {
		acService = &accountService{repo: account.NewRepository(db)}
	})
	return acService
}
