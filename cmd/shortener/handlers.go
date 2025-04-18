package main

import (
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strings"
)

var InMemoryStorage = make(map[string]string)

func rootHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case `GET`:
		linkID := strings.TrimPrefix(req.RequestURI, "/")
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
	case `POST`:
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}
		originalLink := string(bodyBytes)
		if len(originalLink) == 0 {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		linkID := uuid.New().String()
		shortLink := fmt.Sprintf("http://%s/%s", req.Host, linkID)
		InMemoryStorage[linkID] = originalLink
		res.WriteHeader(http.StatusCreated)
		res.Header().Set("Content-Type", "text/plain")
		_, err = res.Write([]byte(shortLink))
		if err != nil {
			panic(err)
		}
	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}

}
