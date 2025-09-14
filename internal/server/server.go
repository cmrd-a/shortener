package server

import (
	"github.com/cmrd-a/shortener/internal/server/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Server represents the HTTP server with its router and middleware configuration.
type Server struct {
	Router *chi.Mux
}

// NewServer creates a new Server instance with configured middleware and routes.
func NewServer(log *zap.Logger, service Servicer, trustedSubnet string) *Server {
	s := &Server{chi.NewRouter()}
	s.Router.Use(
		middleware.RequestResponseLogger(log),
		middleware.CheckContentType,
		middleware.DecompressRequest,
		middleware.CompressResponse,
		middleware.UpsertAuthCookie(log),
	)

	s.Router.Post("/", AddLinkHandler(service))
	s.Router.Get("/{linkId}", GetLinkHandler(service))
	s.Router.Get("/ping", PingHandler(service))

	s.Router.Post("/api/shorten", ShortenHandler(service))
	s.Router.Post("/api/shorten/batch", ShortenBatchHandler(service))
	s.Router.Get("/api/user/urls", GetUserURLsHandler(service))
	s.Router.Delete("/api/user/urls", DeleteUserURLsHandler(service))

	s.Router.Route("/api/internal", func(r chi.Router) {
		r.Use(middleware.TrustedSubnet(trustedSubnet))
		r.Get("/stats", StatsHandler(service))
	})

	return s
}
