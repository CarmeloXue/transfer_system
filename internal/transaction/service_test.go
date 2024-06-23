package transaction

import (
	"context"
	"fmt"
	"main/common/db/testutils"
	"main/common/utils"
	"main/internal/account"
	"main/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type transactionServiceSuite struct {
	suite.Suite
	mockDB *gorm.DB
}

func (s *transactionServiceSuite) SetupTest() {
	s.mockDB, _ = testutils.SetupTestDB()
	s.mockDB.AutoMigrate(model.Account{})
	s.mockDB.AutoMigrate(model.FundMovement{})
	s.mockDB.AutoMigrate(model.Transaction{})

	accouts := []model.Account{
		{
			AccountID: 1,
			Balance:   200,
		},
		{
			AccountID: 2,
			Balance:   100,
		},
	}

	testutils.PrepareData[model.Account](s.mockDB, accouts)
}

func (s *transactionServiceSuite) TeardownTest() {
	sqlDB, _ := s.mockDB.DB()
	sqlDB.Close()
}

func (s *transactionServiceSuite) newMockService() Service {
	return NewService(NewRepository(s.mockDB), account.NewTCCService(s.mockDB), account.NewRepository(s.mockDB))
}

func (s *transactionServiceSuite) Test_CreateTransaction_Happyflow() {
	var (
		req = CreateTransactionRequest{
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               "199.9",
		}
		ctx     = context.Background()
		service = s.newMockService()
	)

	trx, err := service.CreateTransaction(ctx, req)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), trx)

	assert.Equal(s.T(), req.SourceAccountID, trx.SourceAccountID)
	assert.Equal(s.T(), req.DestinationAccountID, trx.DestinationAccountID)
	assert.Equal(s.T(), req.Amount, fmt.Sprint(trx.Amount))
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
	assert.EqualError(s.T(), utils.ErrAmountInvalidFormat, err.Error())
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

func TestTransactionService(t *testing.T) {
	suite.Run(t, &transactionServiceSuite{})
}
