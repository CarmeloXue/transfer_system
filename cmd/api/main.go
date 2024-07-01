package main

import (
	"main/api"
	"main/internal/common/config"
	"main/tools/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

func main() {

	log.Init()
	defer log.Cleanup()
	config.Init()

	r := gin.Default()

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	logger := log.GetLogger()
	logger.Info("Server started")

	api.NewRouter(r)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	// Start the server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Sugar().Error("listen", "err", err)
		}
	}()
	logger.Info("Server started on :8080")
	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

}
