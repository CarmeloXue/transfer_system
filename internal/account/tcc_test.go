package account

import (
	"context"
	"main/common/db/testutils"
	"testing"

	. "main/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type tccSuite struct {
	suite.Suite
	mockDB     *gorm.DB
	repository AccountRepository
}

func (s *tccSuite) SetupTest() {
	s.mockDB, _ = testutils.SetupTestDB()
	s.mockDB.AutoMigrate(FundMovement{})
	s.mockDB.AutoMigrate(Account{})
	s.repository = NewRepository(s.mockDB)
}

func (s *tccSuite) TearDownTest() {
	sqlDB, _ := s.mockDB.DB()
	sqlDB.Close()
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
		accounts = []Account{
			{
				AccountID: 1,
				Balance:   100.0,
			},
			{
				AccountID: 2,
				Balance:   100.0,
			},
		}
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               100.0,
		}
		err error
	)
	s.prepareAccounts(accounts)

	err = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	assert.NoError(s.T(), err)

	fm, err := s.repository.GetFundMovement(ctx, FundMovement{
		TransactionID:    trx.TransactionID,
		FundMovementType: string(FMPayment),
	})
	assert.NoError(s.T(), err)
	s.validateFundMovement(fm, trx, FMPayment)
}

func (s *tccSuite) Test_Try_Insufficient_Should_ReturnError() {
	var (
		accounts = []Account{
			{
				AccountID: 1,
				Balance:   100.0,
			},
			{
				AccountID: 2,
				Balance:   100.0,
			},
		}
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               200.0,
		}
		err error
	)
	s.prepareAccounts(accounts)

	err = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	assert.EqualError(s.T(), ErrInsufficientBalance, err.Error())
}

func (s *tccSuite) Test_Try_MultipleCall_Should_OnlyProceedOnce_ReturnOK() {
	var (
		accounts = []Account{
			{
				AccountID: 1,
				Balance:   100.0,
			},
			{
				AccountID: 2,
				Balance:   100.0,
			},
		}
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               100.0,
		}
		err error
	)

	s.prepareAccounts(accounts)

	err = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	assert.NoError(s.T(), err)

	err = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	assert.NoError(s.T(), err)
}

func (s *tccSuite) Test_Confirm_MultipleCall_Should_OnlyProceedOnce_ReturnOK() {
	var (
		accounts = []Account{
			{
				AccountID: 1,
				Balance:   100.0,
			},
			{
				AccountID: 2,
				Balance:   100.0,
			},
		}
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               100.0,
		}
		err error
	)

	s.prepareAccounts(accounts)

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

	accounts = []Account{
		{
			AccountID: 1,
			Balance:   0,
		},
		{
			AccountID: 2,
			Balance:   200.0,
		},
	}

	s.validateAccounts(ctx, accounts)
}

func (s *tccSuite) Test_Cancel_MultipleCall_Should_OnlyProceedOnce_ReturnOK() {
	var (
		accounts = []Account{
			{
				AccountID: 1,
				Balance:   100.0,
			},
			{
				AccountID: 2,
				Balance:   100.0,
			},
		}
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               100.0,
		}
		err error
	)

	s.prepareAccounts(accounts)

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

	accounts = []Account{
		{
			AccountID: 1,
			Balance:   100,
		},
		{
			AccountID: 2,
			Balance:   100.0,
		},
	}

	s.validateAccounts(ctx, accounts)
}

func (s *tccSuite) Test_TryConfirm_HappyFlow() {
	var (
		accounts = []Account{
			{
				AccountID: 1,
				Balance:   100.0,
			},
			{
				AccountID: 2,
				Balance:   100.0,
			},
		}
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               100.0,
		}
		err error
	)
	s.prepareAccounts(accounts)

	err = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	assert.NoError(s.T(), err)
	err = tcc.Confirm(ctx, trx.TransactionID)
	assert.NoError(s.T(), err)

	payment, err := s.repository.GetFundMovement(ctx, FundMovement{
		TransactionID:    trx.TransactionID,
		FundMovementType: string(FMPayment),
	})
	assert.NoError(s.T(), err)
	s.validateFundMovement(payment, trx, FMPayment)

	paymentRecieved, err := s.repository.GetFundMovement(ctx, FundMovement{
		TransactionID:    trx.TransactionID,
		FundMovementType: string(FMPaymentReceived),
	})
	assert.NoError(s.T(), err)
	s.validateFundMovement(paymentRecieved, trx, FMPaymentReceived)

	expectedAccs := []Account{
		{
			AccountID: 1,
			Balance:   0.0,
		},
		{
			AccountID: 2,
			Balance:   200.0,
		},
	}
	s.validateAccounts(ctx, expectedAccs)
}

func (s *tccSuite) Test_TryCancel_HappyFlow() {
	var (
		accounts = []Account{
			{
				AccountID: 1,
				Balance:   100.0,
			},
			{
				AccountID: 2,
				Balance:   100.0,
			},
		}
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               100.0,
		}
		err error
	)
	s.prepareAccounts(accounts)

	err = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	assert.NoError(s.T(), err)
	err = tcc.Cancel(ctx, trx.TransactionID)
	assert.NoError(s.T(), err)

	payment, err := s.repository.GetFundMovement(ctx, FundMovement{
		TransactionID:    trx.TransactionID,
		FundMovementType: string(FMPayment),
	})
	assert.NoError(s.T(), err)
	s.validateFundMovement(payment, trx, FMPayment)

	refund, err := s.repository.GetFundMovement(ctx, FundMovement{
		TransactionID:    trx.TransactionID,
		FundMovementType: string(FMPaymentRefund),
	})
	assert.NoError(s.T(), err)
	s.validateFundMovement(refund, trx, FMPaymentRefund)
	expectedAccs := []Account{
		{
			AccountID: 1,
			Balance:   100.0,
		},
		{
			AccountID: 2,
			Balance:   100.0,
		},
	}
	s.validateAccounts(ctx, expectedAccs)
}

func (s *tccSuite) Test_EmptyCancel_ShouldSuccess() {
	var (
		accounts = []Account{
			{
				AccountID: 1,
				Balance:   100.0,
			},
			{
				AccountID: 2,
				Balance:   100.0,
			},
		}
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               100.0,
		}
		err error
	)
	s.prepareAccounts(accounts)
	err = tcc.Cancel(ctx, trx.TransactionID)
	assert.NoError(s.T(), err)

	s.validateAccounts(ctx, accounts)
}

func (s *tccSuite) Test_Try_Cancel_Confirm() {
	var (
		accounts = []Account{
			{
				AccountID: 1,
				Balance:   100.0,
			},
			{
				AccountID: 2,
				Balance:   100.0,
			},
		}
		tcc = NewTCCService(s.mockDB)
		ctx = context.Background()
		trx = Transaction{
			TransactionID:        "123",
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               100.0,
		}
		err error
	)
	s.prepareAccounts(accounts)

	err = tcc.Try(ctx, trx.TransactionID, trx.SourceAccountID, trx.DestinationAccountID, trx.Amount)
	assert.NoError(s.T(), err)
	err = tcc.Cancel(ctx, trx.TransactionID)
	assert.NoError(s.T(), err)
	err = tcc.Confirm(ctx, trx.TransactionID)
	assert.NoError(s.T(), err)

	payment, err := s.repository.GetFundMovement(ctx, FundMovement{
		TransactionID:    trx.TransactionID,
		FundMovementType: string(FMPayment),
	})
	assert.NoError(s.T(), err)
	s.validateFundMovement(payment, trx, FMPayment)

	refund, err := s.repository.GetFundMovement(ctx, FundMovement{
		TransactionID:    trx.TransactionID,
		FundMovementType: string(FMPaymentRefund),
	})
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), refund)
	s.validateFundMovement(refund, trx, FMPaymentRefund)

	_, err = s.repository.GetFundMovement(ctx, FundMovement{
		TransactionID:    trx.TransactionID,
		FundMovementType: string(FMPaymentReceived),
	})
	assert.EqualError(s.T(), gorm.ErrRecordNotFound, err.Error())

	s.validateAccounts(ctx, accounts)
}

func (s *tccSuite) validateFundMovement(fm *FundMovement, trx Transaction, fmType FundMovementType) {
	assert.Equal(s.T(), trx.TransactionID, fm.TransactionID)
	assert.Equal(s.T(), trx.SourceAccountID, fm.SourceAccountID)
	assert.Equal(s.T(), trx.Amount, fm.Amount)
	assert.Equal(s.T(), fm.FundMovementType, string(fmType))
}

func (s *tccSuite) validateAccounts(ctx context.Context, expectAccountStatus []Account) {
	for _, acc := range expectAccountStatus {
		resp, err := s.repository.GetAccountByID(ctx, acc.AccountID)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), resp.Balance, acc.Balance)
	}
}

func TestTCCSuite(t *testing.T) {
	suite.Run(t, &tccSuite{})
}
