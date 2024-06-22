package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	baseAccountUrl = "/api/v1/account"
)

func RegisterAccountHanders(r *gin.Engine) {
	r.GET(baseAccountUrl, func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "wellcome to Get account",
		})
	})

	r.POST(baseAccountUrl, func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "wellcome to Post account",
		})
	})
}
