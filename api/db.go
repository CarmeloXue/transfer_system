package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterDBApi(r *gin.Engine) {
	r.GET("/api/db", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "From db",
		})
	})
}
