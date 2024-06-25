package transaction

import (
	"context"
	"fmt"
	"main/common/db/testutils"
	"main/common/utils"
	"main/internal/account"
	tcctestutils "main/internal/account/testutils"

	"main/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type transactionServiceSuite struct {
	suite.Suite
	accountDB     *gorm.DB
	transactionDB *gorm.DB
}

func (s *transactionServiceSuite) SetupTest() {
	s.accountDB, _ = testutils.SetupTestDB()
	s.transactionDB, _ = testutils.SetupTestDB()

	_ = s.accountDB.AutoMigrate(model.Account{})
	_ = s.accountDB.AutoMigrate(model.FundMovement{})
	_ = s.transactionDB.AutoMigrate(model.Transaction{})

	accouts := []model.Account{
		{
			AccountID: 1,
			Balance:   10000000,
		},
		{
			AccountID: 2,
			Balance:   10000000,
		},
	}

	testutils.PrepareData[model.Account](s.accountDB, accouts)
}

func (s *transactionServiceSuite) TeardownTest() {
	sqlDB, _ := s.accountDB.DB()
	sqlDB.Close()

	sqlDB, _ = s.transactionDB.DB()
	sqlDB.Close()
}

func (s *transactionServiceSuite) newMockService() Service {
	return NewService(NewRepository(s.transactionDB), tcctestutils.NewMockTCC(account.NewTCCService(s.accountDB), false, false, false), account.NewRepository(s.accountDB))
}

func (s *transactionServiceSuite) newMockServiceWithTCCTimeout(try, confirm, cancel bool) Service {
	return NewService(NewRepository(s.transactionDB), tcctestutils.NewMockTCC(account.NewTCCService(s.accountDB), try, confirm,
		cancel), account.NewRepository(s.accountDB))
}

func (s *transactionServiceSuite) Test_CreateTransaction_Happyflow() {
	var (
		req = CreateTransactionRequest{
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               "9.0",
		}
		ctx     = context.Background()
		service = s.newMockService()
	)

	trx, err := service.CreateTransaction(ctx, req)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), trx)

	assert.Equal(s.T(), req.SourceAccountID, trx.SourceAccountID)
	assert.Equal(s.T(), req.DestinationAccountID, trx.DestinationAccountID)
	inflatedValue, _ := utils.ParseString(req.Amount)
	assert.Equal(s.T(), inflatedValue, trx.Amount)
}

func (s *transactionServiceSuite) Test_CreateTransaction_InvalidAmount_ShouldReturnError() {
	var (
		req = CreateTransactionRequest{
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               "invalid",
		}
		ctx     = context.Background()
		service = s.newMockService()
	)

	_, err := service.CreateTransaction(ctx, req)
	assert.Error(s.T(), err, err.Error())
}

func (s *transactionServiceSuite) Test_CreateTransaction_InvalidSource_ShouldReturnError() {
	var (
		req = CreateTransactionRequest{
			SourceAccountID:      3,
			DestinationAccountID: 2,
			Amount:               "1",
		}
		ctx     = context.Background()
		service = s.newMockService()
	)

	_, err := service.CreateTransaction(ctx, req)
	assert.ErrorContains(s.T(), err, "invalid sender")
}

func (s *transactionServiceSuite) Test_CreateTransaction_InvalidDestination_ShouldReturnError() {
	var (
		req = CreateTransactionRequest{
			SourceAccountID:      1,
			DestinationAccountID: 4,
			Amount:               "1",
		}
		ctx     = context.Background()
		service = s.newMockService()
	)

	_, err := service.CreateTransaction(ctx, req)
	assert.ErrorContains(s.T(), err, "invalid reciever")
}

func (s *transactionServiceSuite) Test_CreateTransaction_InSufficientBalance_ShouldReturnError() {
	var (
		req = CreateTransactionRequest{
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               "10000000",
		}
		ctx     = context.Background()
		service = s.newMockService()
	)

	_, err := service.CreateTransaction(ctx, req)
	assert.ErrorContains(s.T(), err, "insufficient balance")
}

func (s *transactionServiceSuite) Test_Multiple_Create_Happyflow() {
	var (
		req1To2Amount1 = CreateTransactionRequest{
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               "1",
		}
		req2To1Amount2 = CreateTransactionRequest{
			SourceAccountID:      2,
			DestinationAccountID: 1,
			Amount:               "2",
		}
		ctx     = context.Background()
		service = s.newMockService()
	)

	for i := 0; i < 5; i++ {
		_, _ = service.CreateTransaction(ctx, req1To2Amount1)
		_, _ = service.CreateTransaction(ctx, req2To1Amount2)

	}

	s.validateAccounts(ctx, []model.Account{
		{
			AccountID: 1,
			Balance:   15000000,
		},
		{
			AccountID: 2,
			Balance:   5000000,
		},
	})
}

func (s *transactionServiceSuite) Test_TryTimeout_EmptyCancel_TransactionStatus_ShouldBeFailed() {
	var (
		req = CreateTransactionRequest{
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               "9.0",
		}
		ctx     = context.Background()
		service = s.newMockServiceWithTCCTimeout(true, false, false)
	)

	trx, err := service.CreateTransaction(ctx, req)
	assert.EqualError(s.T(), context.DeadlineExceeded, err.Error())
	assert.NotNil(s.T(), trx)

	assert.Equal(s.T(), req.SourceAccountID, trx.SourceAccountID)
	assert.Equal(s.T(), req.DestinationAccountID, trx.DestinationAccountID)
	inflatedValue, _ := utils.ParseString(req.Amount)
	assert.Equal(s.T(), inflatedValue, trx.Amount)
	assert.Equal(s.T(), model.Failed, trx.TransactionStatus)

}

func (s *transactionServiceSuite) validateAccounts(ctx context.Context, expectAccountStatus []model.Account) {
	accRepo := account.NewRepository(s.accountDB)
	for _, acc := range expectAccountStatus {
		resp, err := accRepo.GetAccountByID(ctx, acc.AccountID)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), acc.Balance, resp.Balance, fmt.Sprintf("mismatch amount for account %v", acc.AccountID))
	}
}

func TestTransactionService(t *testing.T) {
	suite.Run(t, &transactionServiceSuite{})
}
