package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/domain"
)

// ProductService contains business logic for the Product entity.
type ProductService struct {
	repo   domain.ProductRepository
	logger *zap.Logger
}

func NewProductService(repo domain.ProductRepository, logger *zap.Logger) *ProductService {
	return &ProductService{repo: repo, logger: logger}
}

// CreateProduct creates a new product with optional fields and status.
func (s *ProductService) CreateProduct(
	ctx context.Context,
	sku, name string,
	price float64,
	stock int,
	description, imageUrl *string,
	status string,
) (*domain.Product, error) {
	// Allow stock=0 for new products (override generated validator)
	if sku == "" {
		return nil, domain.ErrInvalidProductSku
	}
	if name == "" {
		return nil, domain.ErrInvalidProductName
	}
	if price <= 0 {
		return nil, domain.ErrInvalidProductPrice
	}

	now := time.Now().UTC()
	e := &domain.Product{
		Sku:       sku,
		Name:      name,
		Price:     price,
		Stock:     stock,
		CreatedAt: now,
		UpdatedAt: now,
	}
	e.ID = uuid.New()

	if description != nil {
		e.Description = description
	}
	if imageUrl != nil {
		e.ImageUrl = imageUrl
	}

	if status != "" {
		e.Status = domain.ProductStatus(status)
	} else {
		e.Status = domain.ProductStatusDraft
	}

	if err := s.repo.CreateIfNotExists(ctx, e); err != nil {
		if errors.Is(err, domain.ErrProductDuplicated) {
			return nil, err
		}
		s.logger.Error("failed to create product", zap.Error(err))
		return nil, fmt.Errorf("creating product: %w", err)
	}

	s.logger.Info("product created", zap.String("id", e.ID.String()))
	return e, nil
}

// GetByID finds a product by ID.
func (s *ProductService) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding product: %w", err)
	}
	return e, nil
}

// ListPublished returns only published products.
func (s *ProductService) ListPublished(ctx context.Context) ([]*domain.Product, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing products: %w", err)
	}
	var result []*domain.Product
	for _, p := range all {
		if p.Status == domain.ProductStatusPublished {
			result = append(result, p)
		}
	}
	return result, nil
}

// ListProducts returns all products regardless of status.
func (s *ProductService) ListProducts(ctx context.Context) ([]*domain.Product, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing products: %w", err)
	}
	return items, nil
}

// UpdateProduct applies partial updates to an existing product.
func (s *ProductService) UpdateProduct(
	ctx context.Context,
	id string,
	name *string,
	price *float64,
	stock *int,
	description *string,
	imageUrl *string,
	status *string,
) (*domain.Product, error) {
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding product: %w", err)
	}

	if name != nil {
		e.Name = *name
	}
	if price != nil {
		e.Price = *price
	}
	if stock != nil {
		e.Stock = *stock
	}
	if description != nil {
		e.Description = description
	}
	if imageUrl != nil {
		e.ImageUrl = imageUrl
	}
	if status != nil {
		e.Status = domain.ProductStatus(*status)
	}
	e.UpdatedAt = time.Now().UTC()

	if err := s.repo.Save(ctx, e); err != nil {
		return nil, fmt.Errorf("saving product: %w", err)
	}
	return e, nil
}
