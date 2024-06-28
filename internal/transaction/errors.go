package transaction

import (
	"main/common/response"
	"main/internal/account"
)

var createTransactionErrorMapping = map[error]*response.ExternalResponse{
	account.ErrInsufficientBalance: {
		Code:    400,
		Message: "Insufficient Balance",
	},
	account.ErrExceedingMaxAmount: {
		Code:    400,
		Message: "Reciever Exceeding Maximum Balance Limit",
	},
	ErrSameAccountTransactions: {
		Code:    400,
		Message: "Transfer to Same Account is Not Allowed",
	},
	account.ErrFailedToLoadUser: {
		Code:    400,
		Message: "Sender/Reciever ID Not Found",
	},
	errInvalidParams: {
		Code:    400,
		Message: "Invalid Parameters",
	},
}
