package service

type SvcURL struct {
	ShortURL      string
	OriginalURL   string
	CorrelationID string
	UserID        int64
}
