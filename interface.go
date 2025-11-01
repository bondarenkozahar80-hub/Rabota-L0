package service

import (
	"context"
	"order-service/internal/domain"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	GetOrderByUID(ctx context.Context, orderUID string) (*domain.Order, error)
	ProcessOrderMessage(ctx context.Context, message []byte) error
}
