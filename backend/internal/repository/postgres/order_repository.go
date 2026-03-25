package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JoX23/go-without-magic/internal/config"
	"github.com/JoX23/go-without-magic/internal/domain"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(cfg config.DatabaseConfig) (*OrderRepository, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parsing database DSN: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.MaxOpenConns)
	poolCfg.MinConns = int32(cfg.MaxIdleConns)

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return &OrderRepository{pool: pool}, nil
}

func (r *OrderRepository) CreateIfNotExists(ctx context.Context, e *domain.Order) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO orders (id, user_id, status, total, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		e.ID.String(), e.UserId, e.Status, e.Total, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating order: %w", err)
	}
	return nil
}

func (r *OrderRepository) Save(ctx context.Context, e *domain.Order) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO orders (id, user_id, status, total, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		e.ID.String(), e.UserId, e.Status, e.Total, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting order: %w", err)
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
		result = append(result, e)
	}
	return result, rows.Err()
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
		items = append(items, e)
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
