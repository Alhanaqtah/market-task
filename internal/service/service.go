package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"market/internal/models"
)

type Storage interface {
	OrdersGroupedByShelves(ctx context.Context, orderNums []int64) (map[string][]models.Product, error)
}

type Service struct {
	storage Storage
}

func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) OrdersGroupedByShelves(orderNums string) (map[string][]models.Product, error) {
	const op = "service.OrdersGroupedByShelves"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var orderIDs []int64

	strs := strings.Split(orderNums, ",")

	for _, n := range strs {
		num, err := strconv.Atoi(string(n))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		orderIDs = append(orderIDs, int64(num))
	}

	shelves, err := s.storage.OrdersGroupedByShelves(ctx, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return shelves, nil
}
