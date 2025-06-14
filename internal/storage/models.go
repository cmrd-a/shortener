package storage

//go:generate easyjson -all models.go

type StoredURL struct {
	ShortID     string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      int64  `json:"user_id"`
	IsDeleted   bool   `json:"is_deleted"`
}
