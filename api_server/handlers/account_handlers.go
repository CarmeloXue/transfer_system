package handlers

import (
	"fmt"
	"main/common/db"
	"main/common/log"
	"main/common/response"
	"main/service/account"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	baseAccountUrl = "/api/v1/account"
)

func RegisterAccountHanders(r *gin.Engine) {
	r.POST(baseAccountUrl, func(ctx *gin.Context) {
		var req account.CreateAccountRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			response.ErrorParam(ctx, err.Error())
			return
		}

		db, err := db.GetAccountDBClient()
		if err != nil {
			response.ErrorServer(ctx)
			return
		}

		resp, err := account.NewAccountService(db).CreateAccount(ctx, req)
		if err != nil {
			response.ErrorServerWithErrorMessage(ctx, err)
			return
		}

		response.Ok(ctx, resp)
	})

	r.GET(fmt.Sprintf("%s/:account_id", baseAccountUrl), func(ctx *gin.Context) {
		var req account.QueryAccountRequest
		if err := ctx.ShouldBindUri(&req); err != nil {
			response.ErrorParam(ctx, err.Error())
			return
		}

		db, err := db.GetAccountDBClient()
		if err != nil {
			response.ErrorServer(ctx)
			return
		}

		resp, err := account.NewAccountService(db).QueryAccount(ctx, req)
		log.GetLogger().Info(fmt.Sprintf("query acount resp: %v, err %v\n", resp, err))
		if err != nil {
			if err.Error() == gorm.ErrRecordNotFound.Error() {
				response.ErrorNotFound(ctx)
			} else {
				response.ErrorServer(ctx)
			}
			return
		}

		response.Ok(ctx, resp)
	})
}
