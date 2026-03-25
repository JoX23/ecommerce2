package domain

import "errors"

// Errores de dominio para User — tipados para que el handler los mapee
// correctamente a códigos HTTP o gRPC.
//
// REGLA: estos errores representan casos de negocio, NO errores técnicos.
var (
	ErrUserNotFound            = errors.New("user not found")
	ErrUserDuplicated          = errors.New("user already exists")
	ErrInvalidUserEmail        = errors.New("email cannot be empty")
	ErrInvalidUserName         = errors.New("name cannot be empty")
	ErrInvalidUserPasswordHash = errors.New("password_hash cannot be empty")
)
