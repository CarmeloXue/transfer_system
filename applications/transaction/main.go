package main

import (
	"fmt"
	"main/common/log"
	"main/handlers"
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
	logger.Info("Transaction Server started")
	router := gin.Default()

	handlers.RegisterTransactionHandlers(router)

	srv := &http.Server{
		Addr:    ":8081",
		Handler: r,
	}

	// Start the server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(fmt.Sprintf("listen: %s\n", err))
		}
	}()
	logger.Info("Transaction Server started on :8081")
	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down Transaction server...")

}
