package main

import (
	"fmt"
	"main/common/db"
	"main/common/log"
	"main/internal/account"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

func main() {

	log.Init()
	defer log.Cleanup()

	r := gin.Default()

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	logger := log.GetLogger()
	logger.Info("Server started")
	accoundDB, err := db.GetAccountDBClient()
	if err != nil {
		panic("cannot connect to account database")
	}
	accountService := account.NewService(accoundDB)

	// Initialize Handlers
	accountHandler := account.NewHandler(accountService)

	api := r.Group("/api/v1")
	{
		api.POST("/accounts", accountHandler.CreateAccount)
		api.GET("/accounts/:id", accountHandler.QueryAccount)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Start the server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(fmt.Sprintf("listen: %s\n", err))
		}
	}()
	logger.Info("Server started on :8080")
	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

}