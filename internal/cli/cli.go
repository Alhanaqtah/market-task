package cli

import (
	"fmt"
	"log"
	"os"
	"strings"

	"market/internal/models"
)

type Service interface {
	OrdersGroupedByShelves(orderNums string) (map[string][]models.Product, error)
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

	orders := os.Args[1]

	shelves, err := c.service.OrdersGroupedByShelves(orders)
	if err != nil {
		log.Printf("%s: %w", op, err)
	}

	/* 	for _, p := range shelves {
	   		slices.Reverse(p)
	   	}
	*/
	fmt.Printf("Страница сборки заказов %s\n\n", orders)
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
