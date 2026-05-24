package repository

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	codeAlphabet        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codeLength          = 7
	maxCollisionRetries = 5
	pgUniqueViolation   = "23505"
)

var _ Repository = (*Postgres)(nil)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) CreateURL(ctx context.Context, originalURL string) (string, error) {
	for i := 0; i < maxCollisionRetries; i++ {
		code, err := generateCode(codeLength)
		if err != nil {
			return "", err
		}

		_, err = p.pool.Exec(ctx,
			`INSERT INTO urls (code, original) VALUES ($1, $2)`,
			code, originalURL,
		)
		if err == nil {
			return code, nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			continue
		}
		return "", err
	}
	return "", ErrCodeCollision
}

func (p *Postgres) GetURL(ctx context.Context, code string) (string, error) {
	var original string
	err := p.pool.QueryRow(ctx,
		`SELECT original FROM urls WHERE code = $1`, code,
	).Scan(&original)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return original, err
}

func (p *Postgres) DeleteURL(ctx context.Context, code string) error {
	tag, err := p.pool.Exec(ctx, `DELETE FROM urls WHERE code = $1`, code)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) RecordClick(ctx context.Context, code, ip, userAgent string) error {
	tag, err := p.pool.Exec(ctx, `
		INSERT INTO clicks (url_id, ip, user_agent)
		SELECT id, NULLIF($2, '')::inet, NULLIF($3, '')
		FROM urls WHERE code = $1
	`, code, ip, userAgent)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func generateCode(n int) (string, error) {
	limit := big.NewInt(int64(len(codeAlphabet)))
	buf := make([]byte, n)
	for i := range buf {
		idx, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", err
		}
		buf[i] = codeAlphabet[idx.Int64()]
	}
	return string(buf), nil
}
