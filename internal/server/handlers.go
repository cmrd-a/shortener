package server

import (
	"fmt"
	"github.com/cmrd-a/shortener/internal/config"
	"github.com/cmrd-a/shortener/internal/storage"
	"github.com/cmrd-a/shortener/link"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

var linkService = link.NewService(storage.NewInMemoryRepository())

func AddLinkHandler(res http.ResponseWriter, req *http.Request) {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	originalLink := string(bodyBytes)
	if len(originalLink) == 0 {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	linkID, err := linkService.Add(originalLink)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	shortLink := fmt.Sprintf("%s/%s", config.BaseURL, linkID)
	res.WriteHeader(http.StatusCreated)
	res.Header().Set("Content-Type", "text/plain")
	_, err = res.Write([]byte(shortLink))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetLinkHandler(res http.ResponseWriter, req *http.Request) {
	linkID := chi.URLParam(req, "linkId")
	if len(linkID) == 0 {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	originalLink, err := linkService.Get(linkID)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	http.Redirect(res, req, originalLink, http.StatusTemporaryRedirect)
}
