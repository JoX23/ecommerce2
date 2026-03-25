package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JoX23/go-without-magic/internal/domain"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) CreateIfNotExists(ctx context.Context, e *domain.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx,
		`INSERT INTO orders (id, user_id, status, total, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		e.ID.String(), e.UserId, e.Status, e.Total, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating order: %w", err)
	}

	for _, item := range e.Items {
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (id, order_id, product_id, product_name, qty, unit_price, subtotal)
			 VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6)`,
			e.ID.String(), item.ProductId.String(), item.ProductName, item.Qty, item.UnitPrice, item.Subtotal,
		)
		if err != nil {
			return fmt.Errorf("creating order item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

func (r *OrderRepository) Save(ctx context.Context, e *domain.Order) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO orders (id, user_id, status, total, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (id) DO UPDATE SET
		     status = EXCLUDED.status,
		     total = EXCLUDED.total,
		     updated_at = EXCLUDED.updated_at`,
		e.ID.String(), e.UserId, e.Status, e.Total, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upserting order: %w", err)
	}
	return nil
}

func (r *OrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, status, total, created_at, updated_at
		 FROM orders WHERE id = $1`,
		id,
	)
	e, err := scanOrder(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying by id: %w", err)
	}

	items, err := r.findItemsByOrderID(ctx, e.ID.String())
	if err != nil {
		return nil, err
	}
	e.Items = items

	return e, nil
}

func (r *OrderRepository) FindByUserId(ctx context.Context, userid uuid.UUID) ([]*domain.Order, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, status, total, created_at, updated_at
		 FROM orders WHERE user_id = $1`,
		userid,
	)
	if err != nil {
		return nil, fmt.Errorf("querying by user_id: %w", err)
	}
	defer rows.Close()

	var result []*domain.Order
	for rows.Next() {
		e, err := scanOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		items, err := r.findItemsByOrderID(ctx, e.ID.String())
		if err != nil {
			return nil, err
		}
		e.Items = items
		result = append(result, e)
	}
	return result, rows.Err()
}

func (r *OrderRepository) FindByUserIdPaginated(ctx context.Context, userid uuid.UUID, p domain.PaginationParams) ([]*domain.Order, int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM orders WHERE user_id = $1`,
		userid,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting orders by user_id: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, status, total, created_at, updated_at
		 FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userid, p.Limit, p.Offset(),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("querying by user_id paginated: %w", err)
	}
	defer rows.Close()

	var result []*domain.Order
	for rows.Next() {
		e, err := scanOrder(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning row: %w", err)
		}
		items, err := r.findItemsByOrderID(ctx, e.ID.String())
		if err != nil {
			return nil, 0, err
		}
		e.Items = items
		result = append(result, e)
	}
	return result, total, rows.Err()
}

func (r *OrderRepository) List(ctx context.Context) ([]*domain.Order, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, status, total, created_at, updated_at
		 FROM orders ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing orders: %w", err)
	}
	defer rows.Close()

	var items []*domain.Order
	for rows.Next() {
		e, err := scanOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		orderItems, err := r.findItemsByOrderID(ctx, e.ID.String())
		if err != nil {
			return nil, err
		}
		e.Items = orderItems
		items = append(items, e)
	}
	return items, rows.Err()
}

// findItemsByOrderID retrieves all order_items for a given order ID.
func (r *OrderRepository) findItemsByOrderID(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT product_id, product_name, qty, unit_price, subtotal
		 FROM order_items WHERE order_id = $1`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying order items: %w", err)
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		var productIDStr string
		if err := rows.Scan(&productIDStr, &item.ProductName, &item.Qty, &item.UnitPrice, &item.Subtotal); err != nil {
			return nil, fmt.Errorf("scanning order item: %w", err)
		}
		pid, err := uuid.Parse(productIDStr)
		if err != nil {
			return nil, fmt.Errorf("parsing product uuid: %w", err)
		}
		item.ProductId = pid
		items = append(items, item)
	}
	return items, rows.Err()
}

type orderScanner interface {
	Scan(dest ...any) error
}

func scanOrder(s orderScanner) (*domain.Order, error) {
	var e domain.Order
	var idStr string

	if err := s.Scan(&idStr, &e.UserId, &e.Status, &e.Total, &e.CreatedAt, &e.UpdatedAt); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("parsing uuid: %w", err)
	}
	e.ID = id

	return &e, nil
}
