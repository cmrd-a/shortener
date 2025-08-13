package storage

//go:generate easyjson -all models.go

// StoredURL represents a URL record stored in the repository with all its metadata.
type StoredURL struct {
	ShortID     string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      int64  `json:"user_id"`
	IsDeleted   bool   `json:"is_deleted"`
}

// URLForDelete represents a URL deletion request containing the short ID and user ID.
type URLForDelete struct {
	ShortID string
	UserID  int64
}
