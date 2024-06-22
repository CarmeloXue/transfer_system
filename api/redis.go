package api

import (
	"context"
	"fmt"
	"main/common/log"
	"main/common/redis"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRedisAPI(r *gin.Engine) {
	// Simple handler
	r.POST("/api/nihao", func(ctx *gin.Context) {
		rdb := redis.GetRedisClient()
		cmd := rdb.Set(context.TODO(), "A", "a", 0)
		if cmd.Err() != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": cmd.Err().Error(),
			})
		} else {
			ctx.JSON(http.StatusOK, gin.H{
				"message": "success",
			})
		}
	})

	r.GET("/api/nihao/get", func(ctx *gin.Context) {
		rdb := redis.GetRedisClient()
		cmd := rdb.Get(context.TODO(), "A")
		if cmd.Err() != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": cmd.Err().Error(),
			})
		} else {
			result, _ := cmd.Result()
			log.GetLogger().Info(fmt.Sprintf("result: %v", result))
			ctx.JSON(http.StatusOK, gin.H{
				"message": result,
			})

		}
	})
}
