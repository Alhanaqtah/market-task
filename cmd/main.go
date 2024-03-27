package main

import (
	"log"
	"market/internal/cli"
	"market/internal/service"
	"market/internal/storage/postgres"
)

const connStr = "postgres://postgres:postgres@localhost:5432/market"

func main() {
	storage, err := postgres.New(connStr)
	if err != nil {
		log.Fatalf("failed to init storage connection: %w", err)
	}

	service := service.New(storage)

	cli := cli.New(service)

	cli.Run()
}
