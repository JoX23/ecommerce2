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
// Orden de operaciones:
// 1. Para cada item: buscar producto, validar status y stock → si falla, error sin modificar nada
// 2. Crear el objeto order en memoria
// 3. Persistir la orden
// 4. Solo si la orden se guardó: reducir stock de cada producto
func (s *OrderService) CreateOrder(ctx context.Context, userid uuid.UUID, itemRequests []domain.OrderItemRequest) (*domain.Order, error) {
	if userid == (uuid.UUID{}) {
		return nil, domain.ErrInvalidOrderUserId
	}
	if len(itemRequests) == 0 {
		return nil, fmt.Errorf("order must have at least one item")
	}

	// Paso 1: Validar todos los productos sin modificar nada todavía.
	type resolvedItem struct {
		product *domain.Product
		request domain.OrderItemRequest
	}
	resolved := make([]resolvedItem, 0, len(itemRequests))
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

		if product.Status != domain.ProductStatusPublished {
			return nil, fmt.Errorf("product %s is not available for purchase (status: %s)", product.Sku, product.Status)
		}

		if product.Stock < ir.Qty {
			return nil, fmt.Errorf("insufficient stock for product %s", product.Sku)
		}

		subtotal := product.Price * float64(ir.Qty)
		items = append(items, domain.OrderItem{
			ProductId:   product.ID,
			ProductName: product.Name,
			Qty:         ir.Qty,
			UnitPrice:   product.Price,
			Subtotal:    subtotal,
		})
		total += subtotal
		resolved = append(resolved, resolvedItem{product: product, request: ir})
	}

	// Paso 2: Crear el objeto order en memoria.
	order, err := domain.NewOrder(userid, total, items)
	if err != nil {
		return nil, err
	}

	// Paso 3: Persistir la orden.
	if err := s.repo.CreateIfNotExists(ctx, order); err != nil {
		if errors.Is(err, domain.ErrOrderDuplicated) {
			return nil, err
		}
		s.logger.Error("failed to create order", zap.Error(err))
		return nil, fmt.Errorf("creating order: %w", err)
	}

	// Paso 4: Solo si la orden se guardó, reducir el stock.
	for _, ri := range resolved {
		ri.product.Stock -= ri.request.Qty
		if err := s.productRepo.Save(ctx, ri.product); err != nil {
			s.logger.Error("failed to update product stock after order creation",
				zap.String("order_id", order.ID.String()),
				zap.String("product_id", ri.product.ID.String()),
				zap.Error(err),
			)
			return nil, fmt.Errorf("updating product stock: %w", err)
		}
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

// ListByUser retorna orders for a specific user (no pagination, legacy).
func (s *OrderService) ListByUser(ctx context.Context, userid uuid.UUID) ([]*domain.Order, error) {
	items, err := s.repo.FindByUserId(ctx, userid)
	if err != nil {
		return nil, fmt.Errorf("listing orders by user: %w", err)
	}
	return items, nil
}

// ListByUserPaginated retorna orders de un usuario con paginación.
func (s *OrderService) ListByUserPaginated(ctx context.Context, userid uuid.UUID, params domain.PaginationParams) (*domain.PaginatedResult[*domain.Order], error) {
	items, total, err := s.repo.FindByUserIdPaginated(ctx, userid, params)
	if err != nil {
		return nil, fmt.Errorf("listing orders by user: %w", err)
	}

	totalPages := 0
	if params.Limit > 0 {
		totalPages = (total + params.Limit - 1) / params.Limit
	}

	if items == nil {
		items = []*domain.Order{}
	}

	return &domain.PaginatedResult[*domain.Order]{
		Data:       items,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}
