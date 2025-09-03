package service

import (
	"WBTech_L0/internal/model"
	"WBTech_L0/internal/repository"
	"github.com/google/uuid"
)

type Order interface {
	Insert(order model.Order) error
	GetOrderByID(uuid uuid.UUID) (model.Order, error)
}

type Service struct {
	Order
}

func NewService(repo *repository.Repository) *Service {
	return &Service{
		Order: NewOrderService(repo.Order),
	}
}
