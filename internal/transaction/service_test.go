package transaction

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type transactionServiceSuite struct {
	suite.Suite
}

func (s *transactionServiceSuite) SetupTest() {}

func (s *transactionServiceSuite) TeardownTest() {}

func TestTransactionService(t *testing.T) {
	suite.Run(t, &transactionServiceSuite{})
}
