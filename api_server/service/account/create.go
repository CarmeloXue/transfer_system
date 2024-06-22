package account

import (
	"context"
	"main/common/log"
	"main/common/utils"
	"main/models/account"
)

type (
	CreateAccountRequest struct {
		AccountID      uint64 `json:"account_id" binding:"required"`
		InitialBalance string `json:"initial_balance"`
	}

	// CreateAccountResponse represents the JSON response body structure
	CreateAccountResponse struct {
		AccountID uint64 `json:"account_id"`
		Balance   string `json:"balance"`
	}
)

func (s *accountService) CreateAccount(ctx context.Context, req CreateAccountRequest) (CreateAccountResponse, error) {
	floatValue, err := utils.ParseFloat64String(req.InitialBalance)
	if err != nil {
		log.GetLogger().Error(err.Error())
		return CreateAccountResponse{}, err
	}

	err = s.repo.CreateAccount(&account.Account{
		AccountID: int(req.AccountID),
		Balance:   floatValue,
	})
	if err != nil {
		log.GetLogger().Error(err.Error())
		return CreateAccountResponse{}, err
	}

	return CreateAccountResponse{
		AccountID: req.AccountID,
		Balance:   req.InitialBalance,
	}, nil
}
