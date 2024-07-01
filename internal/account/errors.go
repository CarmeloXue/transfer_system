package account

import (
	"main/internal/model/account"
	"main/tools/response"

	"gorm.io/gorm"
)

var createHandlerErrors = map[error]*response.ExternalResponse{
	account.ErrInternalDuplicatedAccount: {
		Code:    409,
		Message: "Duplicated Account ID",
	},
	account.ErrInvalidRequest: {
		Code:    400,
		Message: "Invalid Request",
	},
}

var queryHandlerErrors = map[error]*response.ExternalResponse{
	gorm.ErrRecordNotFound: {
		Code:    404,
		Message: "Account ID not found",
	},
}
