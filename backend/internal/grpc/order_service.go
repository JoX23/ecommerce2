//go:build ignore
// +build ignore

// NOTE: Este archivo requiere que internal/grpc/proto/order.proto
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

// OrderServiceServerImpl implementa pb.OrderServiceServer
// usando la capa de servicio de dominio existente.
type OrderServiceServerImpl struct {
	svc    *service.OrderService
	logger *zap.Logger
}

func NewOrderServiceServer(svc *service.OrderService, logger *zap.Logger) *OrderServiceServerImpl {
	return &OrderServiceServerImpl{svc: svc, logger: logger}
}

func (s *OrderServiceServerImpl) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	e, err := s.svc.CreateOrder(ctx, req.UserId, req.Total)
	if err != nil {
		return nil, ToGRPCError(err)
	}
	return &pb.CreateOrderResponse{
		Order: asProtoOrder(e),
	}, nil
}

func (s *OrderServiceServerImpl) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	e, err := s.svc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, ToGRPCError(err)
	}
	return &pb.GetOrderResponse{
		Order: asProtoOrder(e),
	}, nil
}

func (s *OrderServiceServerImpl) ListOrders(ctx context.Context, _ *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	items, err := s.svc.ListOrders(ctx)
	if err != nil {
		return nil, ToGRPCError(err)
	}

	out := make([]*pb.Order, 0, len(items))
	for _, e := range items {
		out = append(out, asProtoOrder(e))
	}

	return &pb.ListOrdersResponse{
		Orders: out,
	}, nil
}

func asProtoOrder(e *domain.Order) *pb.Order {
	if e == nil {
		return nil
	}
	return &pb.Order{
		Id:        e.ID.String(),
		UserId:    e.UserId,
		Status:    string(e.Status),
		Total:     e.Total,
		CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
