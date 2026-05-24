package repository

import (
	"context"
	"errors"
)

var (
	ErrNotFound      = errors.New("url not found")
	ErrCodeCollision = errors.New("code collision")
)

type Repository interface {
	CreateURL(ctx context.Context, originalURL string) (string, error)
	GetURL(ctx context.Context, code string) (string, error)
	DeleteURL(ctx context.Context, code string) error
	RecordClick(ctx context.Context, code, ip, userAgent string) error
	Close()
}
