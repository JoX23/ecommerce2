package domain

import (
	"time"

	"github.com/google/uuid"
)

// User es la entidad central del dominio.
// NO depende de ningún framework, ORM ni capa de transporte.
type User struct {
	ID           uuid.UUID
	Email        string
	Name         string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser es el único constructor válido.
// Garantiza que la entidad siempre nace en estado válido.
func NewUser(email string, name string, passwordhash string) (*User, error) {
	if email == "" {
		return nil, ErrInvalidUserEmail
	}
	if name == "" {
		return nil, ErrInvalidUserName
	}
	if passwordhash == "" {
		return nil, ErrInvalidUserPasswordHash
	}

	now := time.Now().UTC()

	return &User{
		ID:           uuid.New(),
		Email:        email,
		Name:         name,
		PasswordHash: passwordhash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}
