package postgres

import (
	"context"
	"fmt"
	"log"

	"market/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	conn *pgxpool.Pool
}

func New(connStr string) (*Storage, error) {
	const op = "storage.postgres.New"

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	/* _, err = conn.Exec(context.Background(),
		`
	CREATE TYPE order_status AS ENUM ('active', 'completed', 'pending');

	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		surname VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL,
		address VARCHAR(255) NOT NULL
	);

	CREATE TABLE IF NOT EXISTS shelves (
		id SERIAL PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		zone VARCHAR(255) NOT NULL
	);

	CREATE TABLE IF NOT EXISTS products (
		id SERIAL PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		category VARCHAR(255) NOT NULL,
		vendor VARCHAR(255) NOT NULL,
		price FLOAT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL,
		status order_status NOT NULL,
		created_at TIMESTAMP NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS order_products (
		id SERIAL PRIMARY KEY,
		order_id INT NOT NULL,
		product_id INT NOT NULL,
		quantity INT NOT NULL,
		FOREIGN KEY (order_id) REFERENCES orders(id),
		FOREIGN KEY (product_id) REFERENCES products(id)
	);

	CREATE TABLE IF NOT EXISTS products_shelves (
		id SERIAL PRIMARY KEY,
		product_id INT NOT NULL,
		shelve_id INT NOT NULL,
		is_main BOOLEAN NOT NULL,
		FOREIGN KEY (product_id) REFERENCES products(id),
		FOREIGN KEY (shelve_id) REFERENCES shelves(id)
	);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	*/
	return &Storage{
		conn: pool,
	}, nil
}

func (s *Storage) OrdersGroupedByShelves(ctx context.Context, orderIDs []int64) (map[string][]models.Product, error) {
	const op = "storage.postgres.OrdersGroupedByShelves"

	shelvesWithOrders := make(map[string][]models.Product)

	for _, orderID := range orderIDs {
		products, err := s.getProductsInOrder(ctx, orderID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		for _, product := range products {
			if _, ok := shelvesWithOrders[product.MainShelf]; !ok {
				shelvesWithOrders[product.MainShelf] = make([]models.Product, 0)
			}

			shelvesWithOrders[product.MainShelf] = append(shelvesWithOrders[product.MainShelf], product)
		}
	}

	return shelvesWithOrders, nil
}

func (s *Storage) getProductsInOrder(ctx context.Context, orderID int64) ([]models.Product, error) {
	const op = "storage.postgres.getProductsInOrder"

	var products []models.Product

	rows, err := s.conn.Query(ctx, `SELECT product_id, quantity FROM order_products WHERE order_id=$1`, orderID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var productID, quantity int64
		if err := rows.Scan(&productID, &quantity); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		product, err := s.getProductInfo(ctx, productID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		product.Count = quantity

		product.OrderID = orderID

		product.ID = productID

		products = append(products, product)
	}

	return products, nil
}

func (s *Storage) getProductInfo(ctx context.Context, productID int64) (models.Product, error) {
	const op = "storage.postgres.getProductInfo"

	var product models.Product

	err := s.conn.QueryRow(ctx, `SELECT title FROM products WHERE id=$1`, productID).Scan(&product.Title)
	if err != nil {
		return product, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := s.conn.Query(ctx, `SELECT shelve_id, is_main FROM products_shelves WHERE product_id=$1`, productID)
	if err != nil {
		return product, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var shelveID int64
		var isMain bool
		if err := rows.Scan(&shelveID, &isMain); err != nil {
			return product, fmt.Errorf("%s: %w", op, err)
		}

		if isMain {
			product.MainShelf = s.shelveNameByID(ctx, shelveID)
		} else {
			product.NonMainShelves = append(product.NonMainShelves, s.shelveNameByID(ctx, shelveID))
		}
	}

	return product, nil
}

func (s *Storage) shelveNameByID(ctx context.Context, shelveID int64) string {
	const op = "storage.postgres.shelveNameByID"

	var title string
	err := s.conn.QueryRow(ctx, `SELECT title FROM shelves WHERE id=$1`, shelveID).Scan(&title)
	if err != nil {
		log.Printf("%s: %v\n", op, err)
	}

	return title
}
