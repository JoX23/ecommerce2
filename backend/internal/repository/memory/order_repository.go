package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/JoX23/go-without-magic/internal/domain"
)

// OrderRepository es una implementación en memoria del dominio.
type OrderRepository struct {
	mu   sync.RWMutex
	byID map[string]*domain.Order
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		byID: make(map[string]*domain.Order),
	}
}

func copyOrder(o *domain.Order) *domain.Order {
	if o == nil {
		return nil
	}
	c := *o // shallow copy del struct
	// copiar el slice de items para evitar aliasing
	if o.Items != nil {
		c.Items = make([]domain.OrderItem, len(o.Items))
		for i, item := range o.Items {
			c.Items[i] = item // OrderItem es un struct (valor), no puntero
		}
	}
	return &c
}

func (r *OrderRepository) CreateIfNotExists(_ context.Context, e *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	stored := copyOrder(e)
	r.byID[e.ID.String()] = stored
	return nil
}

func (r *OrderRepository) Save(_ context.Context, e *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	stored := copyOrder(e)
	r.byID[e.ID.String()] = stored
	return nil
}

func (r *OrderRepository) FindByID(_ context.Context, id string) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrOrderNotFound
	}
	return copyOrder(e), nil
}

func (r *OrderRepository) FindByUserId(_ context.Context, userid uuid.UUID) ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Order
	for _, e := range r.byID {
		if e.UserId == userid {
			result = append(result, copyOrder(e))
		}
	}
	return result, nil
}

func (r *OrderRepository) List(_ context.Context) ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*domain.Order, 0, len(r.byID))
	for _, e := range r.byID {
		items = append(items, copyOrder(e))
	}
	return items, nil
}
