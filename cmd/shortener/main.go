package main

import "net/http"

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", rootHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
