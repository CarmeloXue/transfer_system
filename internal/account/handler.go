package account

import (
	"main/common/response"
	"main/common/utils"
	"main/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateAccount(c *gin.Context) {
	var (
		req         CreateAccountRequest
		returnError *error
	)
	// error handling. Map internal errors to external
	defer func() {
		if returnError != nil {
			response.MapExternalErrors(c, *returnError, createHandlerErrors)
			return
		}
		c.Status(http.StatusCreated)
	}()
	if err := c.ShouldBindJSON(&req); err != nil {
		returnError = &errInvalidRequest
		return
	}
	if err := h.service.CreateAccount(c, req); err != nil {
		returnError = &err
		return
	}

}

func (h *Handler) QueryAccount(c *gin.Context) {
	var (
		req         QueryAccountRequest
		returnError *error
		account     model.Account
	)

	// error handling. Map internal errors to external
	defer func() {
		if returnError != nil {
			response.MapExternalErrors(c, *returnError, queryHandlerErrors)
			return
		}
		displayAccount := QueryResponse{
			AccountID: uint64(account.AccountID),
			Balance:   utils.FormatInt(account.Balance),
		}
		response.Ok(c, displayAccount)
	}()
	if err := c.ShouldBindUri(&req); err != nil {
		returnError = &errInvalidRequest
		return
	}
	account, err := h.service.QueryAccount(c, req)
	if err != nil {
		returnError = &err
		return
	}

}
