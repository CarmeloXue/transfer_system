package transaction

import (
	"main/common/response"
	"main/common/utils"
	"main/internal/account"

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
	utils.ErrNegativeValue: {
		Code:    400,
		Message: "Amount Can Not Be Negative",
	},
	utils.ErrOverflow: {
		Code:    400,
		Message: "Amount Overflow",
	},
	utils.ErrTooManyDigits: {
		Code:    400,
		Message: "Too Many Digits, We Only Support 6 Digits Most",
	},
}
