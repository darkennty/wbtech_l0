package repository

import (
	"WBTech_L0/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Order interface {
	Insert(order model.Order) error
	GetOrderByID(uuid uuid.UUID) (model.Order, error)
	GetOrdersForCache(capacity int) ([]model.Order, error)
}

type Repository struct {
	Order
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Order: NewOrderPostgres(db),
	}
}
