package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"math/rand"
	"net/http"
)

var InMemoryStorage = make(map[string]string)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString() string {
	b := make([]rune, 5)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func AddLinkHandler(res http.ResponseWriter, req *http.Request) {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	originalLink := string(bodyBytes)
	if len(originalLink) == 0 {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	linkID := RandString()
	shortLink := fmt.Sprintf("%s/%s", baseURL, linkID)
	InMemoryStorage[linkID] = originalLink
	res.WriteHeader(http.StatusCreated)
	res.Header().Set("Content-Type", "text/plain")
	_, err = res.Write([]byte(shortLink))
	if err != nil {
		panic(err)
	}
}

func GetLinkHandler(res http.ResponseWriter, req *http.Request) {
	linkID := chi.URLParam(req, "linkId")
	if len(linkID) == 0 {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	originalLink, ok := InMemoryStorage[linkID]
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	http.Redirect(res, req, originalLink, http.StatusTemporaryRedirect)
}
