package main

import (
	"flag"
)

var flagRunAddr string
var baseDomain string

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&baseDomain, "b", "http://localhost:8080", "base domain for short links")
	flag.Parse()
}
