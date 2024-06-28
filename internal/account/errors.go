package account

import (
	"errors"
	"main/common/response"

	"gorm.io/gorm"
)

var (
	errInternalDuplicatedAccount = errors.New("duplicated account")
	errInvalidRequest            = errors.New("invalid request")
)

var createHandlerErrors = map[error]*response.ExternalResponse{
	errInternalDuplicatedAccount: {
		Code:    409,
		Message: "Duplicated Account ID",
	},
	errInvalidRequest: {
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
