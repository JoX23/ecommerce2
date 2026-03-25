package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/JoX23/go-without-magic/internal/domain"
)

// UserService contiene SOLO lógica de negocio.
// No sabe nada de HTTP, gRPC, bases de datos ni frameworks.
type UserService struct {
	repo   domain.UserRepository
	logger *zap.Logger
}

func NewUserService(repo domain.UserRepository, logger *zap.Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}

// CreateUser hashea la contraseña con bcrypt y persiste el usuario.
func (s *UserService) CreateUser(ctx context.Context, email string, name string, password string) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	e, err := domain.NewUser(email, name, string(hash))
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateIfNotExists(ctx, e); err != nil {
		if errors.Is(err, domain.ErrUserDuplicated) {
			return nil, err
		}
		s.logger.Error("failed to create user", zap.Error(err))
		return nil, fmt.Errorf("creating user: %w", err)
	}

	s.logger.Info("user created", zap.String("id", e.ID.String()))
	return e, nil
}

// AuthenticateUser busca el usuario por email y compara el hash de la contraseña.
func (s *UserService) AuthenticateUser(ctx context.Context, email string, password string) (*domain.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	e, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(e.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return e, nil
}

// GetByID busca un user por su ID.
func (s *UserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}
	return e, nil
}

// GetByEmail busca un user por su email.
func (s *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	e, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}
	return e, nil
}

// ListUsers retorna todos los registros.
func (s *UserService) ListUsers(ctx context.Context) ([]*domain.User, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	return items, nil
}
