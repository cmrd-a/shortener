package server

import (
	"github.com/cmrd-a/shortener/internal/config"
	"github.com/cmrd-a/shortener/internal/logger"
	"github.com/go-chi/chi/v5"
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
	s.Router.Use(logger.RequestResponseLogger)

	s.Router.Post("/", AddLinkHandler)
	s.Router.Get("/{linkId}", GetLinkHandler)

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
