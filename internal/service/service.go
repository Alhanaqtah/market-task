package service

import (
	"context"
	"fmt"

	"market/internal/models"
)

type Storage interface {
	OrdersGroupedByShelves(ctx context.Context, orderNums []string) (map[string][]models.Product, error)
}

type Service struct {
	storage Storage
}

func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) OrdersGroupedByShelves(orderNums []string) (map[string][]models.Product, error) {
	const op = "service.OrdersGroupedByShelves"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shelves, err := s.storage.OrdersGroupedByShelves(ctx, orderNums)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return shelves, nil
}
