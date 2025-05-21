package server

import (
	"github.com/cmrd-a/shortener/internal/server/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Server struct {
	Router *chi.Mux
}

func NewServer(log *zap.Logger, service Service, dsn string) *Server {
	s := &Server{chi.NewRouter()}
	s.Router.Use(middleware.RequestResponseLogger(log), middleware.DecompressRequest, middleware.CompressResponse)

	s.Router.Post("/", AddLinkHandler(service))
	s.Router.Get("/{linkId}", GetLinkHandler(service))
	s.Router.Post("/api/shorten", ShortenHandler(service))
	s.Router.Get("/ping", PingHandler(dsn))

	return s
}
