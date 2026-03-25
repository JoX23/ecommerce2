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

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(cfg config.DatabaseConfig) (*UserRepository, error) {
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

	return &UserRepository{pool: pool}, nil
}

func (r *UserRepository) CreateIfNotExists(ctx context.Context, e *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, name, password_hash, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		e.ID.String(), e.Email, e.Name, e.PasswordHash, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return nil
}

func (r *UserRepository) Save(ctx context.Context, e *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, name, password_hash, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		e.ID.String(), e.Email, e.Name, e.PasswordHash, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, name, password_hash, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	)
	e, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying by id: %w", err)
	}
	return e, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, name, password_hash, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	)
	e, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying by email: %w", err)
	}
	return e, nil
}

func (r *UserRepository) List(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, email, name, password_hash, created_at, updated_at
		 FROM users ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var items []*domain.User
	for rows.Next() {
		e, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

type userScanner interface {
	Scan(dest ...any) error
}

func scanUser(s userScanner) (*domain.User, error) {
	var e domain.User
	var idStr string

	if err := s.Scan(&idStr, &e.Email, &e.Name, &e.PasswordHash, &e.CreatedAt, &e.UpdatedAt); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("parsing uuid: %w", err)
	}
	e.ID = id

	return &e, nil
}
