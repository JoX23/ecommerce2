package domain

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus representa los valores posibles del campo Status.
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// OrderItem represents a line item within an order.
type OrderItem struct {
	ProductId uuid.UUID `json:"product_id"`
	Qty       int       `json:"qty"`
	UnitPrice float64   `json:"unit_price"`
	Subtotal  float64   `json:"subtotal"`
}

// OrderItemRequest is used when creating an order to specify product and qty.
type OrderItemRequest struct {
	ProductId uuid.UUID
	Qty       int
}

// Order es la entidad central del dominio.
// NO depende de ningún framework, ORM ni capa de transporte.
type Order struct {
	ID        uuid.UUID
	UserId    uuid.UUID
	Status    OrderStatus
	Total     float64
	Items     []OrderItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewOrder es el único constructor válido.
// Garantiza que la entidad siempre nace en estado válido.
func NewOrder(userid uuid.UUID, total float64, items []OrderItem) (*Order, error) {
	if userid == (uuid.UUID{}) {
		return nil, ErrInvalidOrderUserId
	}
	if total < 0 {
		return nil, ErrInvalidOrderTotal
	}

	now := time.Now().UTC()

	return &Order{
		ID:        uuid.New(),
		UserId:    userid,
		Status:    OrderStatus("pending"),
		Total:     total,
		Items:     items,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
