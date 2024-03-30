package models

type Product struct {
	ID             int64
	MainShelf      string
	Title          string
	OrderID        int64
	Count          int64
	NonMainShelves []string
}
