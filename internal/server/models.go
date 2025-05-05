package server

//go:generate easyjson -all models.go
type ShortenRequest struct {
	URL string `json:"url"`
}
type ShortenResponse struct {
	Result string `json:"result"`
}
