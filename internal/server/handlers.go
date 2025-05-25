package server

import (
	"io"
	"net/http"

	"github.com/mailru/easyjson/jlexer"

	"github.com/cmrd-a/shortener/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
)

func AddLinkHandler(service service.Service) func(http.ResponseWriter, *http.Request) {
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
		shortLink, err := service.Shorten(originalLink)
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

func GetLinkHandler(service service.Service) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ID := chi.URLParam(req, "linkId")
		if len(ID) == 0 {
			http.Error(res, "url is empty", http.StatusBadRequest)
			return
		}
		original, err := service.GetOriginal(ID)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		http.Redirect(res, req, original, http.StatusTemporaryRedirect)
	}
}

func ShortenHandler(service service.Service) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(res, "only Content-Type:application/json is supported", http.StatusBadRequest)
			return
		}
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

		shortLink, err := service.Shorten(reqJSON.URL)
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

func ShortenBatchHandler(svc service.Service) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(res, "only Content-Type:application/json is supported", http.StatusBadRequest)
			return
		}
		var reqJSON ShortenBatchRequest
		buffer := make([]byte, req.ContentLength)
		req.Body.Read(buffer)
		lexer := jlexer.Lexer{Data: buffer}
		reqJSON.UnmarshalEasyJSON(&lexer)
		if err := lexer.Error(); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		if len(reqJSON) == 0 {
			http.Error(res, "urls is empty", http.StatusBadRequest)
			return
		}
		mapForSvc := make(map[string]string, len(reqJSON))
		// todo: проверка на уникальность corr id
		for _, reqURL := range reqJSON {
			if reqURL.OriginalURL == "" || reqURL.CorrelationID == "" {
				http.Error(res, "original_url or correlation_id is empty", http.StatusBadRequest)
				return
			}
			mapForSvc[reqURL.CorrelationID] = reqURL.OriginalURL
		}

		shortenMap, err := svc.ShortenBatch(req.Context(), mapForSvc)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		resJSON := make(ShortenBatchResponse, 0, len(shortenMap))
		for corrID, short := range shortenMap {
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

func PingHandler(svc service.Service) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		err := svc.Ping(req.Context())
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}
