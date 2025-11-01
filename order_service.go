package service

import (
	"context"
	"encoding/json"
	"fmt"
	"order-service/internal/domain"
	"order-service/internal/repository"

	"github.com/go-playground/validator/v10"
)

type orderService struct {
	orderRepo repository.OrderRepository
	cache     repository.Cache
	validator *validator.Validate
}

func NewOrderService(orderRepo repository.OrderRepository, cache repository.Cache) OrderService {
	return &orderService{
		orderRepo: orderRepo,
		cache:     cache,
		validator: validator.New(),
	}
}

func (s *orderService) CreateOrder(ctx context.Context, order *domain.Order) error {
	// Validate order
	if err := s.validator.Struct(order); err != nil {
		return fmt.Errorf("order validation failed: %w", err)
	}

	// Save to database
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return fmt.Errorf("failed to save order to database: %w", err)
	}

	// Cache the order
	s.cache.Set(order.OrderUID, order)
	return nil
}

func (s *orderService) GetOrderByUID(ctx context.Context, orderUID string) (*domain.Order, error) {
	// Try to get from cache first
	if order, exists := s.cache.Get(orderUID); exists {
		return order, nil
	}

	// If not in cache, get from database
	order, err := s.orderRepo.GetByUID(ctx, orderUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order from database: %w", err)
	}

	// Cache the order for future requests
	s.cache.Set(orderUID, order)
	return order, nil
}

func (s *orderService) ProcessOrderMessage(ctx context.Context, message []byte) error {
	var order domain.Order
	if err := json.Unmarshal(message, &order); err != nil {
		return fmt.Errorf("failed to unmarshal order message: %w", err)
	}

	return s.CreateOrder(ctx, &order)
}
