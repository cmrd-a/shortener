package main

import (
	"fmt"
	"github.com/cmrd-a/shortener/internal/config"
	"github.com/cmrd-a/shortener/internal/server"
	"net/http"
)

func main() {
	config.ParseFlags()

	s := server.CreateNewServer()
	s.MountHandlers()
	fmt.Printf("Starting server at %s...\n", config.ServerAddress)
	err := http.ListenAndServe(config.ServerAddress, s.Router)
	if err != nil {
		panic(err)
	}

}
