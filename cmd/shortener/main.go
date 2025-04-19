package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

type Server struct {
	Router *chi.Mux
	// Db, config can be added here
}

func CreateNewServer() *Server {
	s := &Server{}
	s.Router = chi.NewRouter()
	return s
}

func (s *Server) MountHandlers() {
	// Mount all Middleware here
	s.Router.Use(middleware.Logger)

	// Mount all handlers here
	s.Router.Post("/", AddLinkHandler)
	s.Router.Get("/{linkId}", GetLinkHandler)

}

func main() {
	parseFlags()

	s := CreateNewServer()
	s.MountHandlers()

	err := http.ListenAndServe(flagRunAddr, s.Router)
	if err != nil {
		panic(err)
	}

}
