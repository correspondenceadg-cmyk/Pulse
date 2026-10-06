package auth

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RefreshToken struct {
	ID          string
	UserID      string
	TokenHash   string
	RotatedFrom *string
	ExpiresAt   time.Time
	RevokedAt   *time.Time
	CreatedAt   time.Time
}