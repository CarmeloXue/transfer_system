package transaction

import (
	"main/common/response"
	"main/internal/account"
	"main/tools/currency"

	"gorm.io/gorm"
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
	gorm.ErrRecordNotFound: {
		Code:    400,
		Message: "Sender/Reciever ID Not Found",
	},
	errInvalidParams: {
		Code:    400,
		Message: "Invalid Parameters",
	},
	currency.ErrNegativeValue: {
		Code:    400,
		Message: "Amount Can Not Be Negative",
	},
	currency.ErrOverflow: {
		Code:    400,
		Message: "Amount Overflow",
	},
	currency.ErrTooManyDigits: {
		Code:    400,
		Message: "Too Many Digits, We Only Support 6 Digits Most",
	},
}
