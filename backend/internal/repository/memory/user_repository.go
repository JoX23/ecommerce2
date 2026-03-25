package memory

import (
	"context"
	"sync"

	"github.com/JoX23/go-without-magic/internal/domain"
)

// UserRepository es una implementación en memoria del dominio.
//
// Usos:
//   - Tests unitarios y de integración (sin base de datos real)
//   - Desarrollo local sin infraestructura
//
// Es seguro para uso concurrente (sync.RWMutex).
type UserRepository struct {
	mu      sync.RWMutex
	byID    map[string]*domain.User
	byEmail map[string]*domain.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		byID:    make(map[string]*domain.User),
		byEmail: make(map[string]*domain.User),
	}
}

// CreateIfNotExists verifica unicidad y crea de forma ATÓMICA.
// Retorna ErrUserDuplicated si ya existe un duplicado.
//
// GARANTÍA DE CONCURRENCIA: Thread-safe, sin ventana entre check y write.
func (r *UserRepository) CreateIfNotExists(_ context.Context, e *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byEmail[e.Email]; exists {
		return domain.ErrUserDuplicated
	}

	r.byID[e.ID.String()] = e
	r.byEmail[e.Email] = e

	return nil
}

func (r *UserRepository) Save(_ context.Context, e *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.byID[e.ID.String()] = e
	r.byEmail[e.Email] = e

	return nil
}

func (r *UserRepository) FindByID(_ context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return e, nil
}

func (r *UserRepository) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, ok := r.byEmail[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return e, nil
}

func (r *UserRepository) List(_ context.Context) ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*domain.User, 0, len(r.byID))
	for _, e := range r.byID {
		items = append(items, e)
	}
	return items, nil
}
