package token

import (
	"time"
)

type Maker interface {
	CreateToken(
		id int64,
		email string,
		role string,
		name string,
		instituteID int32,
		duration time.Duration,
	) (string, *TokenPayload, error)

	VerifyToken(token string) (*TokenPayload, error)
}

type TokenPayload struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	IssuedAt    time.Time `json:"issued_at"`
	InstituteID int32     `json:"institute_id"`
	ExpiredAt   time.Time `json:"expired_at"`
	Name        string    `json:"name"`
}

func NewTokenPayload(id int64, email string, role string, name string, duration time.Duration) (*TokenPayload, error) {
	return &TokenPayload{
		ID:        id,
		Email:     email,
		Role:      role,
		Name:      name,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}, nil
}
