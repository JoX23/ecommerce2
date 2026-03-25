package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/domain"
)

// OrderService contiene SOLO lógica de negocio.
type OrderService struct {
	repo        domain.OrderRepository
	productRepo domain.ProductRepository
	logger      *zap.Logger
}

func NewOrderService(repo domain.OrderRepository, productRepo domain.ProductRepository, logger *zap.Logger) *OrderService {
	return &OrderService{
		repo:        repo,
		productRepo: productRepo,
		logger:      logger,
	}
}

// CreateOrder creates an order from a list of items with product lookups.
func (s *OrderService) CreateOrder(ctx context.Context, userid uuid.UUID, itemRequests []domain.OrderItemRequest) (*domain.Order, error) {
	if userid == (uuid.UUID{}) {
		return nil, domain.ErrInvalidOrderUserId
	}
	if len(itemRequests) == 0 {
		return nil, fmt.Errorf("order must have at least one item")
	}

	var items []domain.OrderItem
	var total float64

	for _, ir := range itemRequests {
		product, err := s.productRepo.FindByID(ctx, ir.ProductId.String())
		if err != nil {
			if errors.Is(err, domain.ErrProductNotFound) {
				return nil, fmt.Errorf("product %s not found", ir.ProductId)
			}
			return nil, fmt.Errorf("looking up product: %w", err)
		}
		if product.Stock < ir.Qty {
			return nil, fmt.Errorf("insufficient stock for product %s", product.Sku)
		}

		subtotal := product.Price * float64(ir.Qty)
		items = append(items, domain.OrderItem{
			ProductId: product.ID,
			Qty:       ir.Qty,
			UnitPrice: product.Price,
			Subtotal:  subtotal,
		})
		total += subtotal

		// Reduce stock
		product.Stock -= ir.Qty
		if err := s.productRepo.Save(ctx, product); err != nil {
			return nil, fmt.Errorf("updating product stock: %w", err)
		}
	}

	order, err := domain.NewOrder(userid, total, items)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateIfNotExists(ctx, order); err != nil {
		if errors.Is(err, domain.ErrOrderDuplicated) {
			return nil, err
		}
		s.logger.Error("failed to create order", zap.Error(err))
		return nil, fmt.Errorf("creating order: %w", err)
	}

	s.logger.Info("order created", zap.String("id", order.ID.String()))
	return order, nil
}

// GetByID busca un order por su ID.
func (s *OrderService) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding order: %w", err)
	}
	return e, nil
}

// ListOrders retorna todos los registros.
func (s *OrderService) ListOrders(ctx context.Context) ([]*domain.Order, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing orders: %w", err)
	}
	return items, nil
}

// ListByUser retorna orders for a specific user.
func (s *OrderService) ListByUser(ctx context.Context, userid uuid.UUID) ([]*domain.Order, error) {
	items, err := s.repo.FindByUserId(ctx, userid)
	if err != nil {
		return nil, fmt.Errorf("listing orders by user: %w", err)
	}
	return items, nil
}
