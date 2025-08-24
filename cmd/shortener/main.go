package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

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

func main() {
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
	s := server.NewServer(zl, svc)
	defer func(Log *zap.Logger) {
		err := Log.Sync()
		if err != nil {
			log.Fatal(err)
		}
	}(zl)
	zl.Info("Running server", zap.String("address", cfg.ServerAddress))
	err = http.ListenAndServe(cfg.ServerAddress, s.Router)
	if err != nil {
		log.Fatal(err)
	}
}
