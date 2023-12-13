package pkg

import (
	"context"
	"sync"
)

const (
	UserIDParamName = "userID"
)

type BatchURLs struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ResultBatchShort struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type UserURL struct {
	Code string `json:"short_url"`
	URL  string `json:"original_url"`
}

type ShortURL struct {
	URL       string `db:"url"`
	IsDeleted bool   `db:"is_deleted"`
}

type APIShortener interface {
	Auth(token string) (int, error)
	GenerateUserID() int
	CreateToken() (int, string, error)
	EncodeURL(ctx context.Context, userID int, url string) (string, bool)
	BatchEncodeURL(ctx context.Context, userID int, batch []BatchURLs) []ResultBatchShort
	DecodeURL(ctx context.Context, shortURL string) (string, error)
	AddShortly(ctx context.Context, UserID int, url, code string)
	GetUserURLs(ctx context.Context, userID int) []UserURL
	DeleteUserURLs(ctx context.Context, userID int, codes []string)
}

type Storable interface {
	sync.Locker
	InitStorage() error
	GetCode(ctx context.Context, url string) (string, bool)
	GetCodeBatch(ctx context.Context, urls []string) map[string]string
	GetURL(ctx context.Context, code string) (ShortURL, bool)
	GetMaxUserID(ctx context.Context) (int, error)
	GetUserURLs(ctx context.Context, userID int) []UserURL
	Insert(ctx context.Context, userID int, url, code string) error
	InsertBatch(ctx context.Context, userID int, batch [][]string) error
	DeleteBatch(ctx context.Context, userID int, batch []string)
	Count() int
	Truncate() error
	Ping() error
	Close()
}

type GeneratorID interface {
	GetID() int
}
