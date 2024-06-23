package transaction

import (
	"main/common/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateTransaction(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorParam(c, err.Error())
		return
	}
	trx, err := h.service.CreateTransaction(c, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"data":    trx,
		})
		return
	}
	response.Ok(c, trx)
}

func (h *Handler) QueryTransaction(c *gin.Context) {
	var req QueryTransactionRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.ErrorParam(c, err.Error())
		return
	}
	trx, err := h.service.QueryTransaction(c, req)
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
