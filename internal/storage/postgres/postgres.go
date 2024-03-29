package postgres

import (
	"context"
	"fmt"
	"strings"

	"market/internal/models"

	"github.com/jackc/pgx/v5"
)

type Storage struct {
	conn *pgx.Conn
}

func New(connStr string) (*Storage, error) {
	const op = "storage.postgres.New"

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = conn.Ping(context.Background())
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
		conn: conn,
	}, nil
}

func (s *Storage) OrdersGroupedByShelves(ctx context.Context, orderNums []string) (map[string][]models.Product, error) {
	const op = "storage.postgres.OrdersGroupedByShelves"

	rows, err := s.conn.Query(ctx, `
	SELECT 
		s.title,
		p.title,
		p.id,
		o.id,
		op.quantity,
		ps.is_main
	FROM 
		orders o
	JOIN 
		order_products op ON o.id = op.order_id
	JOIN 
		products p ON p.id = op.product_id
	JOIN 
		products_shelves ps ON ps.product_id = p.id
	JOIN
		shelves s ON ps.shelve_id = s.id
	WHERE 
		o.id IN (`+strings.Join(orderNums, ",")+`)
	ORDER BY 
		s.title DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	shelves := make(map[string][]models.Product)
	for rows.Next() {
		var shelveTitle, productTitle string
		var productID, orderID, productCount int64
		var isMain bool

		err := rows.Scan(&shelveTitle, &productTitle, &productID, &orderID, &productCount, &isMain)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if isMain {
			if _, ok := shelves[shelveTitle]; !ok {
				shelves[shelveTitle] = make([]models.Product, 0)
			}

			prod := models.Product{
				ID:      productID,
				Title:   productTitle,
				OrderID: orderID,
				Count:   productCount,
			}

			shelves[shelveTitle] = append(shelves[shelveTitle], prod)
		} else {
			for _, p := range shelves {
				for i := 0; i < len(p); i++ {
					if p[i].Title == productTitle {
						p[i].NonMainShelves = append(p[i].NonMainShelves, shelveTitle)
					}
				}
			}
		}
	}

	return shelves, nil
}
