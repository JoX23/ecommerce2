//go:build ignore
// +build ignore

// NOTE: Este archivo requiere que internal/grpc/proto/user.proto
// haya sido compilado con protoc primero. Elimina las líneas build ignore
// una vez que el pb package esté disponible.

package grpc

import (
	"context"

	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/domain"
	"github.com/JoX23/go-without-magic/internal/grpc/pb"
	"github.com/JoX23/go-without-magic/internal/service"
)

// UserServiceServerImpl implementa pb.UserServiceServer
// usando la capa de servicio de dominio existente.
type UserServiceServerImpl struct {
	svc    *service.UserService
	logger *zap.Logger
}

func NewUserServiceServer(svc *service.UserService, logger *zap.Logger) *UserServiceServerImpl {
	return &UserServiceServerImpl{svc: svc, logger: logger}
}

func (s *UserServiceServerImpl) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	e, err := s.svc.CreateUser(ctx, req.Email, req.Name, req.PasswordHash)
	if err != nil {
		return nil, ToGRPCError(err)
	}
	return &pb.CreateUserResponse{
		User: asProtoUser(e),
	}, nil
}

func (s *UserServiceServerImpl) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	e, err := s.svc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, ToGRPCError(err)
	}
	return &pb.GetUserResponse{
		User: asProtoUser(e),
	}, nil
}

func (s *UserServiceServerImpl) ListUsers(ctx context.Context, _ *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	items, err := s.svc.ListUsers(ctx)
	if err != nil {
		return nil, ToGRPCError(err)
	}

	out := make([]*pb.User, 0, len(items))
	for _, e := range items {
		out = append(out, asProtoUser(e))
	}

	return &pb.ListUsersResponse{
		Users: out,
	}, nil
}

func asProtoUser(e *domain.User) *pb.User {
	if e == nil {
		return nil
	}
	return &pb.User{
		Id:           e.ID.String(),
		Email:        e.Email,
		Name:         e.Name,
		PasswordHash: e.PasswordHash,
		CreatedAt:    e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
