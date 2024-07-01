package account

import (
	"context"
	"main/internal/common/db/testutils"
	"testing"

	. "main/internal/model/account"
	trxModel "main/internal/model/transaction"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type tccSuite struct {
	suite.Suite
	mockDB          *gorm.DB
	repository      AccountRepository
	defaultAccounts []Account
}

func (s *tccSuite) SetupTest() {
	s.mockDB, _ = testutils.SetupTestDB()
	_ = s.mockDB.AutoMigrate(FundMovement{})
	_ = s.mockDB.AutoMigrate(Account{})
	s.repository = NewRepository(s.mockDB)

	s.defaultAccounts = []Account{
		{
			AccountID: 1,
			Balance:   100000000, // inflated amount
		},
		{
			AccountID: 2,
			Balance:   100000000,
		},
	}
	s.prepareAccounts(s.defaultAccounts)
}

func (s *tccSuite) TearDownTest() {
	s.mockDB.Exec("DELETE FROM account_tab")
	s.mockDB.Exec("DELETE FROM fund_movement_tab")

}

func (s *tccSuite) prepareAccounts(accounts []Account) {
	for _, acc := range accounts {
		result := s.mockDB.Create(&acc)
		if result.Error != nil {
			panic("failed to seed data")
		}
	}

}

func (s *tccSuite) Test_Try_Happyflow() {
	var (
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = trxModel.Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               100000000,
		}
		err error
	)

	err = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	assert.NoError(s.T(), err)

	err = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	assert.NoError(s.T(), err)

	err = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	assert.NoError(s.T(), err)

	fm, err := s.repository.GetFundMovement(ctx, FundMovement{
		TransactionID: trx.TransactionID,
	})
	assert.NoError(s.T(), err)
	s.validateFundMovement(fm, trx, Tried)
}

func (s *tccSuite) Test_Try_Insufficient_Should_ReturnError() {
	var (
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = trxModel.Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               2000000000000000000,
		}
		err error
	)

	err = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	assert.EqualError(s.T(), ErrInsufficientBalance, err.Error())
}

func (s *tccSuite) Test_Confirm_MultipleCall_Should_OnlyProceedOnce_ReturnOK() {
	var (
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = trxModel.Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               100.0,
		}
		err error
	)

	err = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	assert.NoError(s.T(), err)

	err = tcc.Confirm(ctx, trx.TransactionID)
	assert.NoError(s.T(), err)
	err = tcc.Confirm(ctx, trx.TransactionID)
	assert.NoError(s.T(), err)
	err = tcc.Confirm(ctx, trx.TransactionID)
	assert.NoError(s.T(), err)
	err = tcc.Confirm(ctx, trx.TransactionID)
	assert.NoError(s.T(), err)

	accounts := []Account{
		{
			AccountID: 1,
			Balance:   99999900,
		},
		{
			AccountID: 2,
			Balance:   100000100,
		},
	}

	s.validateAccounts(ctx, accounts)
}

func (s *tccSuite) Test_Cancel_MultipleCall_Should_OnlyProceedOnce_ReturnOK() {
	var (
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = trxModel.Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               100.0,
		}
		err error
	)

	err = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	assert.NoError(s.T(), err)

	err = tcc.Cancel(ctx, trx.TransactionID)
	assert.NoError(s.T(), err)
	err = tcc.Cancel(ctx, trx.TransactionID)
	assert.NoError(s.T(), err)
	err = tcc.Cancel(ctx, trx.TransactionID)
	assert.NoError(s.T(), err)
	err = tcc.Cancel(ctx, trx.TransactionID)
	assert.NoError(s.T(), err)

	s.validateAccounts(ctx, s.defaultAccounts)
}

func (s *tccSuite) Test_TryConfirm_HappyFlow() {
	var (
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = trxModel.Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               100,
		}
		err error
	)

	_ = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	err = tcc.Confirm(ctx, trx.TransactionID)
	assert.NoError(s.T(), err)

	fm, err := s.repository.GetFundMovement(ctx, FundMovement{
		TransactionID: trx.TransactionID,
	})
	assert.NoError(s.T(), err)
	s.validateFundMovement(fm, trx, Confirmed)

	expectedAccs := []Account{
		{
			AccountID: 1,
			Balance:   99999900,
		},
		{
			AccountID: 2,
			Balance:   100000100,
		},
	}
	s.validateAccounts(ctx, expectedAccs)
}

func (s *tccSuite) Test_TryCancel_HappyFlow() {
	var (
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = trxModel.Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               100.0,
		}
		err error
	)

	err = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	assert.NoError(s.T(), err)
	err = tcc.Cancel(ctx, trx.TransactionID)
	assert.NoError(s.T(), err)

	rollback, err := s.repository.GetFundMovement(ctx, FundMovement{
		TransactionID: trx.TransactionID,
	})
	assert.NoError(s.T(), err)
	s.validateFundMovement(rollback, trx, Canceled)

	s.validateAccounts(ctx, s.defaultAccounts)
}

func (s *tccSuite) Test_EmptyCancel_ShouldSuccess() {
	var (
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = trxModel.Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               100.0,
		}
		err error
	)
	err = tcc.Cancel(ctx, trx.TransactionID)
	assert.EqualError(s.T(), ErrEmptyRollback, err.Error())
	var rollback FundMovement
	err = s.mockDB.First(&rollback, FundMovement{
		TransactionID: trx.TransactionID,
	}).Error
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(0), rollback.Amount)
	assert.Equal(s.T(), Canceled, rollback.Stage)

	s.validateAccounts(ctx, s.defaultAccounts)
}

func (s *tccSuite) Test_Try_Cancel_Confirm() {
	var (
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = trxModel.Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               100.0,
		}
		err error
	)

	err = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	assert.NoError(s.T(), err)
	err = tcc.Cancel(ctx, trx.TransactionID)
	assert.NoError(s.T(), err)
	err = tcc.Confirm(ctx, trx.TransactionID)
	assert.EqualError(s.T(), ErrRollbacked, err.Error())

	refund, err := s.repository.GetFundMovement(ctx, FundMovement{
		TransactionID: trx.TransactionID,
	})
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), refund)
	s.validateFundMovement(refund, trx, Canceled)

	s.validateAccounts(ctx, s.defaultAccounts)
}

func (s *tccSuite) Test_Cancel_Try() {
	var (
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = trxModel.Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               100.0,
		}
		err error
	)

	err = tcc.Cancel(ctx, trx.TransactionID)
	assert.EqualError(s.T(), ErrEmptyRollback, err.Error())
	err = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	assert.EqualError(s.T(), ErrRollbacked, err.Error())

	refund, err := s.repository.GetFundMovement(ctx, FundMovement{
		TransactionID: trx.TransactionID,
	})
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), refund)
	assert.Equal(s.T(), int64(0), refund.Amount)
	assert.Equal(s.T(), Canceled, refund.Stage)
	s.validateAccounts(ctx, s.defaultAccounts)
}

func (s *tccSuite) validateFundMovement(fm *FundMovement, trx trxModel.Transaction, stage FundMovementStage) {
	assert.Equal(s.T(), trx.TransactionID, fm.TransactionID, "transaction_id not match")
	assert.Equal(s.T(), trx.SourceAccountID, fm.SourceAccountID, "source_id not match")
	assert.Equal(s.T(), trx.Amount, fm.Amount, "acount not match")
	assert.Equal(s.T(), stage, fm.Stage, "stage not match")
}

func (s *tccSuite) validateAccounts(ctx context.Context, expectAccountStatus []Account) {
	for _, acc := range expectAccountStatus {
		resp, err := s.repository.GetAccountByID(ctx, acc.AccountID)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), acc.Balance, resp.Balance, "balance not match")
	}
}

func TestTCCSuite(t *testing.T) {
	suite.Run(t, &tccSuite{})
}
