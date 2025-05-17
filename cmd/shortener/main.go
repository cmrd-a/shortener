package main

import (
	"log"
	"net/http"

	"github.com/cmrd-a/shortener/internal/config"
	"github.com/cmrd-a/shortener/internal/logger"
	"github.com/cmrd-a/shortener/internal/server"
	"github.com/cmrd-a/shortener/internal/service"
	"github.com/cmrd-a/shortener/internal/storage"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()
	zl := logger.NewLogger(cfg.LogLevel)
	URLService := service.NewURLService(cfg.BaseURL, storage.NewFileRepository(cfg.FileStoragePath))
	s := server.NewServer(zl, URLService)
	defer func(Log *zap.Logger) {
		err := Log.Sync()
		if err != nil {
			log.Fatal(err)
		}
	}(zl)
	zl.Info("Running server", zap.String("address", cfg.ServerAddress))
	err := http.ListenAndServe(cfg.ServerAddress, s.Router)
	if err != nil {
		log.Fatal(err)
	}
}
