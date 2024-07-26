package api

import (
	"main/internal/account"
	"main/internal/common/cache"
	"main/internal/common/db"
	"main/internal/common/queue"
	accModel "main/internal/model/account"
	trxModel "main/internal/model/transaction"
	"main/internal/transaction"

	"github.com/gin-gonic/gin"
)

func NewRouter(r *gin.Engine) {
	accoundDB, err := db.GetAccountDBClient()
	if err != nil {
		panic("cannot connect to account database")
	}
	accountCache, err := cache.NewRedisClient()
	if err != nil {
		panic("cannot connect to account database")
	}
	accountService := account.NewService(accoundDB, accountCache)

	// Initialize Handlers
	accountHandler := account.NewHandler(accountService)

	transactionDB, err := db.GetTransactionDB()
	if err != nil {
		panic("cannot connect to account database")
	}
	messenger, err := queue.NewTransactionProducer(&transaction.TransactionMessageHandler{
		Topic: transaction.TransactionCreationTopic,
	})
	if err != nil {
		panic("cannot connect to producer database")
	}

	transactionHandler := transaction.NewHandler(
		transaction.NewService(trxModel.NewRepository(transactionDB),
			account.NewTCCService(accoundDB),
			accModel.NewRepository(accoundDB),
			messenger),
	)

	api := r.Group("/api/v1")
	api.POST("/accounts", accountHandler.CreateAccount)
	api.GET("/accounts/:account_id", accountHandler.QueryAccount)
	api.POST("/transactions", transactionHandler.CreateTransaction)
	api.GET("/transactions/:transaction_id", transactionHandler.QueryTransaction)
	api.POST("/transactions/retry", transactionHandler.RetryTransaction)
}
