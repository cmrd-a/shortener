package service

// SvcURL represents a URL record in the service layer containing both short and original URLs.
type SvcURL struct {
	ShortURL    string
	OriginalURL string
	UserID      int64
	IsDeleted   bool
}
