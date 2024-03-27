package cli

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"market/internal/models"
)

type Service interface {
	OrdersGroupedByShelves(orderNums []string) (map[string][]models.Product, error)
}

type Cli struct {
	service Service
}

func New(service Service) *Cli {
	return &Cli{
		service: service,
	}
}

func (c *Cli) Run() {
	const op = "cli.Run"

	var orders []string

	if strings.Join(os.Args[:2], " ") == "go run main.go" {
		orders = append(orders, os.Args[3:]...)
	} else {
		orders = append(orders, os.Args[1:]...)
	}

	shelves, err := c.service.OrdersGroupedByShelves(orders)
	if err != nil {
		log.Printf("%s: %w", op, err)
	}

	for _, p := range shelves {
		slices.Reverse(p)
	}

	fmt.Printf("Страница сборки заказов %s\n\n", strings.Join(orders, ","))
	for shelveTitle, products := range shelves {
		fmt.Printf("===Стеллаж %s\n", shelveTitle)
		for _, product := range products {
			fmt.Printf("%s (id=%d)\n", product.Title, product.ID)
			fmt.Printf("заказ %d, %d шт\n\n", product.OrderID, product.Count)
			if len(product.NonMainShelves) != 0 {
				fmt.Printf("доп стеллаж: %s\n\n", strings.Join(product.NonMainShelves, ","))
			}
		}
	}
}
