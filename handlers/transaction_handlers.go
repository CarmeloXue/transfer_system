package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterTransactionHandlers(r *gin.Engine) {
	r.POST("/api/v1/payment", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "wellcome to payment",
		})
	})
}
