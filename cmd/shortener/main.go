package main

import (
	"github.com/cmrd-a/shortener/internal/config"
	"github.com/cmrd-a/shortener/internal/logger"
	"github.com/cmrd-a/shortener/internal/server"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	config.ParseFlags()
	s := server.CreateNewServer()
	s.Prepare()
	defer func(Log *zap.Logger) {
		err := Log.Sync()
		if err != nil {
			panic(err)
		}
	}(logger.Log)
	logger.Log.Info("Running server", zap.String("address", config.ServerAddress))
	err := http.ListenAndServe(config.ServerAddress, s.Router)
	if err != nil {
		panic(err)
	}
}
