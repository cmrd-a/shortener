package server

// ShortenRequest represents the JSON request body for shortening a single URL.
//
//go:generate easyjson -all models.go
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse represents the JSON response body containing a shortened URL.
type ShortenResponse struct {
	Result string `json:"result"`
}

// ShortenBatchRequest represents a batch request for shortening multiple URLs.
//
//easyjson:json
type ShortenBatchRequest []ShortenBatchRequestItem

// ShortenBatchRequestItem represents a single item in a batch shortening request.
type ShortenBatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ShortenBatchResponse represents a batch response containing multiple shortened URLs.
//
//easyjson:json
type ShortenBatchResponse []ShortenBatchResponseItem

// ShortenBatchResponseItem represents a single item in a batch shortening response.
type ShortenBatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// GetUserURLsResponse represents the response containing all URLs for a specific user.
//
//easyjson:json
type GetUserURLsResponse []GetUserURLsResponseItem

// GetUserURLsResponseItem represents a single URL item in the user's URL list.
type GetUserURLsResponseItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// DeleteUserURLsRequest represents a request to delete multiple URLs for a user.
//
//easyjson:json
type DeleteUserURLsRequest []string
