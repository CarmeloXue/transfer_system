package transaction

import (
	"errors"
	"main/common/response"
	"main/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	service Service
}

var errInvalidParams = errors.New("invalid params")

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateTransaction(c *gin.Context) {
	var (
		req         CreateTransactionRequest
		returnError *error
		err         error
		trx         model.Transaction
	)
	defer func() {
		if returnError != nil {
			response.MapExternalErrors(c, *returnError, createTransactionErrorMapping)
			return
		}
		(&trx).FormatForDisplay()
		response.Ok(c, trx)
	}()
	if err := c.ShouldBindJSON(&req); err != nil {
		returnError = &errInvalidParams
		return
	}
	trx, err = h.service.CreateTransaction(c, req)

	if err != nil {
		returnError = &err
		return
	}

}

func (h *Handler) QueryTransaction(c *gin.Context) {
	var req QueryTransactionRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.ErrorParam(c, err.Error())
		return
	}
	trx, err := h.service.QueryTransaction(c, req)
	(&trx).FormatForDisplay()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.ErrorNotFound(c)
		} else {
			response.ErrorServer(c)
		}
		return
	}
	response.Ok(c, trx)
}

func (h *Handler) RetryTransaction(c *gin.Context) {
	var req QueryTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorParam(c, err.Error())
		return
	}
	trx, err := h.service.RetryTransaction(c, req)
	(&trx).FormatForDisplay()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.ErrorNotFound(c)
		} else {
			response.ErrorServer(c)
		}
		return
	}
	response.Ok(c, trx)
}
