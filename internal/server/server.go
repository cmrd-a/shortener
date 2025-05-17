package server

import (
	"github.com/cmrd-a/shortener/internal/server/middleware"
	"github.com/cmrd-a/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type Server struct {
	Router *chi.Mux
}

func NewServer(log *zap.Logger, service *service.URLService) *Server {
	s := &Server{}
	s.Router = chi.NewRouter()

	s.Router.Use(middleware.RequestResponseLogger(log), middleware.DecompressRequest, chiMiddleware.Compress(5, "text/html", "application/json"), chiMiddleware.AllowContentEncoding("gzip"))

	s.Router.Post("/", AddLinkHandler(service))
	s.Router.Get("/{linkId}", GetLinkHandler(service))
	s.Router.Post("/api/shorten", ShortenHandler(service))
	return s
}
