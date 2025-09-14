package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cmrd-a/shortener/internal/config"
	"github.com/cmrd-a/shortener/internal/logger"
	"github.com/cmrd-a/shortener/internal/server"
	"github.com/cmrd-a/shortener/internal/service"
	"github.com/cmrd-a/shortener/internal/storage"

	"go.uber.org/zap"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func main() {
	printBuildInfo()

	cfg := config.NewConfig(true)
	zl, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Printf("ERROR: failed to initialize logger %s \n", err)
		zl = zap.NewNop()
	}
	ctx := context.Background()
	repo, err := storage.MakeRepository(ctx, cfg)
	if err != nil {
		log.Fatalf("ERROR: failed to initialize repository %s \n", err)
	}
	generator := service.NewShortGenerator()
	svc := service.NewURLService(generator, cfg.BaseURL, repo)
	s := server.NewServer(zl, svc, cfg.TrustedSubnet)
	defer func(Log *zap.Logger) {
		err := Log.Sync()
		if err != nil {
			log.Fatal(err)
		}
	}(zl)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: s.Router,
	}

	// Create a channel to receive shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Start server in a goroutine
	go func() {
		zl.Info("Running server", zap.String("address", cfg.ServerAddress))

		var err error
		if cfg.EnableHTTPS {
			server.GenerateTLS()
			err = httpServer.ListenAndServeTLS("cert.pem", "private_key.pem")
		} else {
			err = httpServer.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			zl.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	zl.Info("Shutdown signal received")

	// Create a timeout context for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the HTTP server gracefully
	zl.Info("Shutting down HTTP server...")
	if err := httpServer.Shutdown(ctx); err != nil {
		zl.Error("Server shutdown failed", zap.Error(err))
	} else {
		zl.Info("Server shutdown completed")
	}

	// Close storage repository if it has a Close method
	if closer, ok := repo.(interface{ Close() error }); ok {
		zl.Info("Closing storage repository...")
		if err := closer.Close(); err != nil {
			zl.Error("Failed to close storage repository", zap.Error(err))
		} else {
			zl.Info("Storage repository closed successfully")
		}
	}

	zl.Info("Application shutdown completed")
}
