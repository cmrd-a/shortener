package server

import (
	"github.com/cmrd-a/shortener/internal/config"
	"github.com/cmrd-a/shortener/internal/logger"
	"github.com/cmrd-a/shortener/internal/server/middleware"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Router *chi.Mux
}

func CreateNewServer() *Server {
	s := &Server{}
	s.Router = chi.NewRouter()
	return s
}

func (s *Server) MountHandlers() {
	s.Router.Use(middleware.RequestResponseLogger, middleware.DecompressRequest, chiMiddleware.Compress(5, "text/html", "application/json"), chiMiddleware.AllowContentEncoding("gzip"))

	s.Router.Post("/", AddLinkHandler)
	s.Router.Get("/{linkId}", GetLinkHandler)
	s.Router.Post("/api/shorten", ShortenHandler)

}
func (s *Server) InitLogger() {
	if err := logger.Initialize(config.LogLevel); err != nil {
		panic(err)
	}
}

func (s *Server) Prepare() {
	s.InitLogger()
	s.MountHandlers()
}
