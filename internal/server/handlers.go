package server

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/cmrd-a/shortener/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
)

type Servicer interface {
	Shorten(string) (string, error)
	ShortenBatch(context.Context, map[string]string) (map[string]string, error)
	GetOriginal(string) (string, error)
	Ping(context.Context) error
}

func AddLinkHandler(svc Servicer) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		originalLink := string(bodyBytes)
		if len(originalLink) == 0 {
			http.Error(res, "url is empty", http.StatusBadRequest)
			return
		}
		shortLink, err := svc.Shorten(originalLink)
		var alreadyExistError *service.OriginalExistError
		if errors.As(err, &alreadyExistError) {
			res.WriteHeader(http.StatusConflict)
			res.Write([]byte(err.Error()))
			return
		}
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusCreated)
		res.Header().Set("Content-Type", "text/plain")
		_, err = res.Write([]byte(shortLink))
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetLinkHandler(svc Servicer) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ID := chi.URLParam(req, "linkId")
		if len(ID) == 0 {
			http.Error(res, "url is empty", http.StatusBadRequest)
			return
		}
		original, err := svc.GetOriginal(ID)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		http.Redirect(res, req, original, http.StatusTemporaryRedirect)
	}
}

func ShortenHandler(svc Servicer) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		reqJSON := &ShortenRequest{}
		err := easyjson.UnmarshalFromReader(req.Body, reqJSON)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		if len(reqJSON.URL) == 0 {
			http.Error(res, "url is empty", http.StatusBadRequest)
			return
		}

		shortLink, err := svc.Shorten(reqJSON.URL)
		var alreadyExistError *service.OriginalExistError
		if errors.As(err, &alreadyExistError) {
			body, err := ShortenResponse{Result: alreadyExistError.Short}.MarshalJSON()
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
			}
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusConflict)

			res.Write(body)
			return
		}
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		resJSON := &ShortenResponse{Result: shortLink}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)

		resBytes, err := resJSON.MarshalJSON()
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = res.Write(resBytes)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func ShortenBatchHandler(svc Servicer) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.ContentLength == 0 {
			res.WriteHeader(http.StatusOK)
			return
		}
		var reqJSON ShortenBatchRequest
		err := easyjson.UnmarshalFromReader(req.Body, &reqJSON)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}
		corrOrig := make(map[string]string, len(reqJSON))
		for _, reqItem := range reqJSON {
			if reqItem.OriginalURL == "" || reqItem.CorrelationID == "" {
				http.Error(res, "original_url or correlation_id is empty", http.StatusBadRequest)
				return
			}
			if _, ok := corrOrig[reqItem.CorrelationID]; ok {
				http.Error(res, "duplicated correlation_id"+reqItem.CorrelationID, http.StatusBadRequest)
			}
			for _, original := range corrOrig {
				if original == reqItem.OriginalURL {
					http.Error(res, "duplicates original_url"+reqItem.OriginalURL, http.StatusBadRequest)
				}
			}
			corrOrig[reqItem.CorrelationID] = reqItem.OriginalURL
		}

		corrShort, err := svc.ShortenBatch(req.Context(), corrOrig)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		resJSON := make(ShortenBatchResponse, 0, len(corrShort))
		for corrID, short := range corrShort {
			item := ShortenBatchResponseItem{CorrelationID: corrID, ShortURL: short}
			resJSON = append(resJSON, item)
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)

		resBytes, err := resJSON.MarshalJSON()
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = res.Write(resBytes)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func PingHandler(svc Servicer) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		err := svc.Ping(req.Context())
		if err != nil {
			http.Error(res, err.Error(), http.StatusServiceUnavailable)
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}
