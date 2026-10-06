package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailTaken        = errors.New("email taken")
	ErrRefreshNotFound   = errors.New("refresh not found")
	ErrRefreshRevoked    = errors.New("refresh revoked")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateUser(ctx context.Context, email, hash, display string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name)
		VALUES (lower($1), $2, $3)
		RETURNING id, email, password_hash, coalesce(display_name,''), role, created_at, updated_at
	`, email, hash, display).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, coalesce(display_name,''), role, created_at, updated_at
		FROM users
		WHERE lower(email) = lower($1)
	`, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, coalesce(display_name,''), role, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) StoreRefresh(ctx context.Context, userID, hash string, expires time.Time, rotatedFrom *string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, rotated_from)
		VALUES ($1, $2, $3, $4)
	`, userID, hash, expires, rotatedFrom)
	return err
}

func (r *Repository) FindRefresh(ctx context.Context, hash string) (*RefreshToken, error) {
	var rt RefreshToken
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, rotated_from, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, hash).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.RotatedFrom, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRefreshNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *Repository) RevokeRefresh(ctx context.Context, hash string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, hash)
	return err
}

func (r *Repository) RevokeUserChain(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now()
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	return err
}

func isUniqueViolation(err error) bool {
	type pgErr interface{ SQLState() string }
	var p pgErr
	if errors.As(err, &p) {
		return p.SQLState() == "23505"
	}
	return false
}