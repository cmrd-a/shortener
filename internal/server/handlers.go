package server

import (
	"fmt"
	"github.com/cmrd-a/shortener/internal/config"
	"github.com/cmrd-a/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

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
	linkID := storage.RandString()
	shortLink := fmt.Sprintf("%s/%s", config.BaseURL, linkID)
	storage.InMemoryStorage[linkID] = originalLink
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
	originalLink, ok := storage.InMemoryStorage[linkID]
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	http.Redirect(res, req, originalLink, http.StatusTemporaryRedirect)
}
