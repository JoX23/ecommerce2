package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/domain"
	"github.com/JoX23/go-without-magic/internal/repository/memory"
	"github.com/JoX23/go-without-magic/internal/service"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

func newUserService(t *testing.T) *service.UserService {
	t.Helper()
	return service.NewUserService(memory.NewUserRepository(), zap.NewNop())
}

func newProductService(t *testing.T) *service.ProductService {
	t.Helper()
	return service.NewProductService(memory.NewProductRepository(), zap.NewNop())
}

type orderServices struct {
	orderSvc   *service.OrderService
	productSvc *service.ProductService
}

func newOrderServices(t *testing.T) orderServices {
	t.Helper()
	productRepo := memory.NewProductRepository()
	orderRepo := memory.NewOrderRepository()
	return orderServices{
		orderSvc:   service.NewOrderService(orderRepo, productRepo, zap.NewNop()),
		productSvc: service.NewProductService(productRepo, zap.NewNop()),
	}
}

// ── UserService tests ─────────────────────────────────────────────────────────

func TestUserService_CreateUser_Success(t *testing.T) {
	svc := newUserService(t)
	ctx := context.Background()

	user, err := svc.CreateUser(ctx, "alice@example.com", "Alice", "password123")

	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", user.Email)
	assert.Equal(t, "Alice", user.Name)
	assert.NotEmpty(t, user.ID)
	assert.False(t, user.CreatedAt.IsZero())
	// Password must be hashed, not stored in plain text
	assert.NotEqual(t, "password123", user.PasswordHash)
	assert.NotEmpty(t, user.PasswordHash)
}

func TestUserService_CreateUser_DuplicateEmail(t *testing.T) {
	svc := newUserService(t)
	ctx := context.Background()

	_, err := svc.CreateUser(ctx, "bob@example.com", "Bob", "password123")
	require.NoError(t, err)

	_, err = svc.CreateUser(ctx, "bob@example.com", "Bob Clone", "password456")

	assert.ErrorIs(t, err, domain.ErrUserDuplicated)
}

func TestUserService_CreateUser_InvalidEmail(t *testing.T) {
	svc := newUserService(t)
	ctx := context.Background()

	_, err := svc.CreateUser(ctx, "", "Alice", "password123")

	assert.ErrorIs(t, err, domain.ErrInvalidUserEmail)
}

func TestUserService_CreateUser_InvalidName(t *testing.T) {
	svc := newUserService(t)
	ctx := context.Background()

	_, err := svc.CreateUser(ctx, "alice@example.com", "", "password123")

	assert.ErrorIs(t, err, domain.ErrInvalidUserName)
}

func TestUserService_AuthenticateUser_Success(t *testing.T) {
	svc := newUserService(t)
	ctx := context.Background()

	created, err := svc.CreateUser(ctx, "charlie@example.com", "Charlie", "securepass")
	require.NoError(t, err)

	authenticated, err := svc.AuthenticateUser(ctx, "charlie@example.com", "securepass")

	require.NoError(t, err)
	assert.Equal(t, created.ID, authenticated.ID)
	assert.Equal(t, created.Email, authenticated.Email)
}

func TestUserService_AuthenticateUser_WrongPassword(t *testing.T) {
	svc := newUserService(t)
	ctx := context.Background()

	_, err := svc.CreateUser(ctx, "dave@example.com", "Dave", "correctpass")
	require.NoError(t, err)

	_, err = svc.AuthenticateUser(ctx, "dave@example.com", "wrongpass")

	assert.Error(t, err)
}

func TestUserService_AuthenticateUser_UserNotFound(t *testing.T) {
	svc := newUserService(t)
	ctx := context.Background()

	_, err := svc.AuthenticateUser(ctx, "nonexistent@example.com", "password")

	assert.Error(t, err)
}

// ── ProductService tests ───────────────────────────────────────────────────────

func TestProductService_CreateProduct_Success(t *testing.T) {
	svc := newProductService(t)
	ctx := context.Background()

	product, err := svc.CreateProduct(ctx, "SKU-001", "Widget", 9.99, 100, nil, nil, "published")

	require.NoError(t, err)
	assert.Equal(t, "SKU-001", product.Sku)
	assert.Equal(t, "Widget", product.Name)
	assert.Equal(t, 9.99, product.Price)
	assert.Equal(t, 100, product.Stock)
	assert.Equal(t, domain.ProductStatusPublished, product.Status)
	assert.NotEmpty(t, product.ID)
}

func TestProductService_CreateProduct_DefaultStatus(t *testing.T) {
	svc := newProductService(t)
	ctx := context.Background()

	product, err := svc.CreateProduct(ctx, "SKU-002", "Gadget", 19.99, 50, nil, nil, "")

	require.NoError(t, err)
	assert.Equal(t, domain.ProductStatusDraft, product.Status)
}

func TestProductService_ListPublished_OnlyShowsPublished(t *testing.T) {
	svc := newProductService(t)
	ctx := context.Background()

	_, err := svc.CreateProduct(ctx, "SKU-PUB", "Published Product", 10.0, 10, nil, nil, "published")
	require.NoError(t, err)

	_, err = svc.CreateProduct(ctx, "SKU-DRF", "Draft Product", 20.0, 5, nil, nil, "draft")
	require.NoError(t, err)

	_, err = svc.CreateProduct(ctx, "SKU-ARC", "Archived Product", 30.0, 0, nil, nil, "archived")
	require.NoError(t, err)

	published, err := svc.ListPublished(ctx)

	require.NoError(t, err)
	require.Len(t, published, 1)
	assert.Equal(t, "SKU-PUB", published[0].Sku)
}

func TestProductService_ListPublished_Empty(t *testing.T) {
	svc := newProductService(t)
	ctx := context.Background()

	_, err := svc.CreateProduct(ctx, "SKU-DRF", "Draft Only", 10.0, 5, nil, nil, "draft")
	require.NoError(t, err)

	published, err := svc.ListPublished(ctx)

	require.NoError(t, err)
	assert.Empty(t, published)
}

func TestProductService_UpdateProduct_Success(t *testing.T) {
	svc := newProductService(t)
	ctx := context.Background()

	product, err := svc.CreateProduct(ctx, "SKU-UPD", "Original Name", 5.0, 10, nil, nil, "draft")
	require.NoError(t, err)

	newName := "Updated Name"
	newPrice := 15.0
	newStatus := "published"

	updated, err := svc.UpdateProduct(ctx, product.ID.String(), &newName, &newPrice, nil, nil, nil, &newStatus)

	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, 15.0, updated.Price)
	assert.Equal(t, domain.ProductStatusPublished, updated.Status)
	// Stock should remain unchanged
	assert.Equal(t, 10, updated.Stock)
}

// ── OrderService tests ─────────────────────────────────────────────────────────

func TestOrderService_CreateOrder_Success_ReducesStock(t *testing.T) {
	svcs := newOrderServices(t)
	ctx := context.Background()

	product, err := svcs.productSvc.CreateProduct(ctx, "SKU-ORD", "Orderable Product", 25.0, 50, nil, nil, "published")
	require.NoError(t, err)

	userID := uuid.New()
	items := []domain.OrderItemRequest{
		{ProductId: product.ID, Qty: 3},
	}

	order, err := svcs.orderSvc.CreateOrder(ctx, userID, items)

	require.NoError(t, err)
	assert.NotEmpty(t, order.ID)
	assert.Equal(t, userID, order.UserId)
	assert.Equal(t, 75.0, order.Total) // 25.0 * 3
	assert.Len(t, order.Items, 1)
	assert.Equal(t, domain.OrderStatusPending, order.Status)

	// Verify stock was reduced
	updated, err := svcs.productSvc.GetByID(ctx, product.ID.String())
	require.NoError(t, err)
	assert.Equal(t, 47, updated.Stock) // 50 - 3
}

func TestOrderService_CreateOrder_InsufficientStock(t *testing.T) {
	svcs := newOrderServices(t)
	ctx := context.Background()

	product, err := svcs.productSvc.CreateProduct(ctx, "SKU-LOW", "Low Stock Product", 10.0, 2, nil, nil, "published")
	require.NoError(t, err)

	userID := uuid.New()
	items := []domain.OrderItemRequest{
		{ProductId: product.ID, Qty: 5}, // more than available
	}

	_, err = svcs.orderSvc.CreateOrder(ctx, userID, items)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient stock")

	// Stock must not have changed
	unchanged, err := svcs.productSvc.GetByID(ctx, product.ID.String())
	require.NoError(t, err)
	assert.Equal(t, 2, unchanged.Stock)
}

func TestOrderService_CreateOrder_ArchivedProduct_Fails(t *testing.T) {
	svcs := newOrderServices(t)
	ctx := context.Background()

	product, err := svcs.productSvc.CreateProduct(ctx, "SKU-ARC2", "Archived Product", 10.0, 10, nil, nil, "archived")
	require.NoError(t, err)

	userID := uuid.New()
	items := []domain.OrderItemRequest{
		{ProductId: product.ID, Qty: 1},
	}

	_, err = svcs.orderSvc.CreateOrder(ctx, userID, items)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available for purchase")
}

func TestOrderService_CreateOrder_DraftProduct_Fails(t *testing.T) {
	svcs := newOrderServices(t)
	ctx := context.Background()

	product, err := svcs.productSvc.CreateProduct(ctx, "SKU-DRF2", "Draft Product", 10.0, 10, nil, nil, "draft")
	require.NoError(t, err)

	userID := uuid.New()
	items := []domain.OrderItemRequest{
		{ProductId: product.ID, Qty: 1},
	}

	_, err = svcs.orderSvc.CreateOrder(ctx, userID, items)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available for purchase")
}

func TestOrderService_CreateOrder_MultipleItems(t *testing.T) {
	svcs := newOrderServices(t)
	ctx := context.Background()

	p1, err := svcs.productSvc.CreateProduct(ctx, "SKU-A", "Product A", 10.0, 20, nil, nil, "published")
	require.NoError(t, err)
	p2, err := svcs.productSvc.CreateProduct(ctx, "SKU-B", "Product B", 30.0, 10, nil, nil, "published")
	require.NoError(t, err)

	userID := uuid.New()
	items := []domain.OrderItemRequest{
		{ProductId: p1.ID, Qty: 2},
		{ProductId: p2.ID, Qty: 1},
	}

	order, err := svcs.orderSvc.CreateOrder(ctx, userID, items)

	require.NoError(t, err)
	assert.Len(t, order.Items, 2)
	assert.Equal(t, 50.0, order.Total) // 10*2 + 30*1

	// Verify stocks reduced
	up1, _ := svcs.productSvc.GetByID(ctx, p1.ID.String())
	up2, _ := svcs.productSvc.GetByID(ctx, p2.ID.String())
	assert.Equal(t, 18, up1.Stock)
	assert.Equal(t, 9, up2.Stock)
}
