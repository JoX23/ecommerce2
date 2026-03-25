package domain

import "context"

// UserRepository define el contrato del puerto de salida.
//
// El dominio define la interfaz; la implementación vive en
// internal/repository/. Esto es el patrón Port & Adapter.
//
// REGLA: esta interfaz NO importa nada fuera del paquete domain.
type UserRepository interface {
	// CreateIfNotExists crea la entidad si no existe un duplicado.
	// Retorna ErrUserDuplicated si ya existe.
	//
	// GARANTÍA: Thread-safe. Operación atómica.
	CreateIfNotExists(ctx context.Context, e *User) error

	// Save crea o actualiza la entidad (incondicionalmente).
	Save(ctx context.Context, e *User) error

	// FindByID busca por ID.
	// Retorna ErrUserNotFound si no existe.
	FindByID(ctx context.Context, id string) (*User, error)

	// FindByEmail busca por Email.
	// Retorna ErrUserNotFound si no existe.
	FindByEmail(ctx context.Context, email string) (*User, error)

	// List retorna todos los registros.
	List(ctx context.Context) ([]*User, error)
}
