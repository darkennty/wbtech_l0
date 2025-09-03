package repository

import (
	"WBTech_L0/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testDir = "../testdata/model.json"
)

func TestInsert(t *testing.T) {
	db, teardown := TestDB(t)
	defer teardown("\"order\"", "delivery", "payment", "item")

	repo := NewRepository(db)
	err := repo.Order.Insert(NewOrder(testDir))
	assert.NoError(t, err)
}

func TestGetOrderByID(t *testing.T) {
	db, teardown := TestDB(t)
	defer teardown("\"order\"", "delivery", "payment", "item")

	repo := NewRepository(db)
	order := NewOrder(testDir)
	err := repo.Order.Insert(order)
	assert.NoError(t, err)

	newOrder, err := repo.GetOrderByID(order.OrderUID)
	assert.NoError(t, err)
	assert.Equal(t, order, newOrder)
}

func TestGetOrdersForCache(t *testing.T) {
	db, teardown := TestDB(t)
	defer teardown("\"order\"", "delivery", "payment", "item")

	repo := NewRepository(db)

	order1 := NewOrder(testDir)
	order2 := NewOrder(testDir)
	order3 := NewOrder(testDir)
	order4 := NewOrder(testDir)
	order5 := NewOrder(testDir)

	repo.Order.Insert(order1)
	repo.Order.Insert(order2)
	repo.Order.Insert(order3)
	repo.Order.Insert(order4)
	repo.Order.Insert(order5)

	orders, err := repo.GetOrdersForCache(5)
	assert.NoError(t, err)

	assert.NotEqual(t, orders, []model.Order{order1, order1, order1, order1, order1})
	assert.Equal(t, orders, []model.Order{order1, order2, order3, order4, order5})
}
