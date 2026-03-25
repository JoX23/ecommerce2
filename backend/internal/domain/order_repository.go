package domain

import (
	"context"

	"github.com/google/uuid"
)

// OrderRepository define el contrato del puerto de salida.
//
// El dominio define la interfaz; la implementación vive en
// internal/repository/. Esto es el patrón Port & Adapter.
//
// REGLA: esta interfaz NO importa nada fuera del paquete domain.
type OrderRepository interface {
	// CreateIfNotExists crea la entidad si no existe un duplicado.
	// Retorna ErrOrderDuplicated si ya existe.
	//
	// GARANTÍA: Thread-safe. Operación atómica.
	CreateIfNotExists(ctx context.Context, e *Order) error

	// Save crea o actualiza la entidad (incondicionalmente).
	Save(ctx context.Context, e *Order) error

	// FindByID busca por ID.
	// Retorna ErrOrderNotFound si no existe.
	FindByID(ctx context.Context, id string) (*Order, error)

	// FindByUserId busca por UserId.
	FindByUserId(ctx context.Context, userid uuid.UUID) ([]*Order, error)

	// FindByUserIdPaginated busca orders de un usuario con paginación.
	// Retorna: items de la página, total de orders del usuario, error.
	FindByUserIdPaginated(ctx context.Context, userid uuid.UUID, p PaginationParams) ([]*Order, int, error)

	// List retorna todos los registros.
	List(ctx context.Context) ([]*Order, error)
}
