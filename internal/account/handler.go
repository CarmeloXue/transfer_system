package account

import (
	"main/common/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateAccount(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorParam(c, err.Error())
		return
	}
	if err := h.service.CreateAccount(c, req); err != nil {
		response.ErrorServer(c)
		return
	}
	c.Status(http.StatusCreated)
}

func (h *Handler) QueryAccount(c *gin.Context) {
	var req QueryAccountRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.ErrorParam(c, err.Error())
		return
	}
	account, err := h.service.QueryAccount(c, req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.ErrorNotFound(c)
			return
		}
		if strings.Contains(err.Error(), "duplicate key") {
			response.ErrorDuplicated(c, err.Error())
		} else {
			response.ErrorServer(c)
		}
		return
	}
	c.JSON(http.StatusOK, account)
}
