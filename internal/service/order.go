package service

import (
	"WBTech_L0/internal/model"
	"WBTech_L0/internal/repository"
	"github.com/google/uuid"
)

type OrderService struct {
	repo repository.Order
}

func NewOrderService(repo repository.Order) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) Insert(order model.Order) error {
	return s.repo.Insert(order)
}

func (s *OrderService) GetOrderByID(uuid uuid.UUID) (model.Order, error) {
	return s.repo.GetOrderByID(uuid)
}
