package account

import (
	"context"
	"fmt"
	"main/models/account"
)

type (
	QueryAccountRequest struct {
		AccountID uint64 `uri:"account_id" json:"account_id" binding:"required"`
	}

	// CreateAccountResponse represents the JSON response body structure
	QueryResponse struct {
		AccountID uint64 `json:"account_id"`
		Balance   string `json:"balance"`
	}
)

func (s *accountService) QueryAccount(ctx context.Context, req QueryAccountRequest) (QueryResponse, error) {
	var (
		acc account.Account
		err error
	)
	if acc, err = s.repo.GetAccountByID(int(req.AccountID)); err != nil {
		return QueryResponse{}, err
	}

	return QueryResponse{
		AccountID: uint64(acc.AccountID),
		Balance:   fmt.Sprint(acc.Balance),
	}, nil
}
