package testutils

import (
	"context"
	"main/internal/account"
)

type mockTCC struct {
	tcc            account.TCC
	tryTimeout     bool
	confirmTimeout bool
	canceltimeout  bool
}

func NewMockTCC(tcc account.TCC, try, confirm, cancel bool) *mockTCC {
	return &mockTCC{tcc: tcc, tryTimeout: try, confirmTimeout: confirm, canceltimeout: cancel}
}

func (s *mockTCC) SetTimeout(try, confirm, cancel bool) {
	s.tryTimeout = try
	s.confirmTimeout = confirm
	s.canceltimeout = cancel
}

func (s *mockTCC) Try(ctx context.Context, transactionID string, sourceAccountID, destinationAccountID int, amount float64) error {
	if s.tryTimeout {
		return context.DeadlineExceeded
	}
	return s.tcc.Try(ctx, transactionID, sourceAccountID, destinationAccountID, amount)
}

func (s *mockTCC) Confirm(ctx context.Context, transactionID string) error {
	if s.confirmTimeout {
		return context.DeadlineExceeded
	}
	return s.tcc.Confirm(ctx, transactionID)
}

func (s *mockTCC) Cancel(ctx context.Context, transactionID string) error {
	if s.canceltimeout {
		return context.DeadlineExceeded
	}
	return s.tcc.Cancel(ctx, transactionID)
}
