package server

//go:generate easyjson -all models.go
type ShortenRequest struct {
	URL string `json:"url"`
}
type ShortenResponse struct {
	Result string `json:"result"`
}

//easyjson:json
type ShortenBatchRequest []ShortenBatchRequestItem

type ShortenBatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

//easyjson:json
type ShortenBatchResponse []ShortenBatchResponseItem
type ShortenBatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
