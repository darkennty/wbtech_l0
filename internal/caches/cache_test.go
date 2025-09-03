package caches

import (
	"WBTech_L0/internal/repository"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testDir = "../testdata/model.json"
)

func TestCache_Set(t *testing.T) {
	db, teardown := repository.TestDB(t)
	defer teardown("\"order\"", "delivery", "payment", "item")

	repo := repository.NewRepository(db)
	testCache := NewCache(repo, 100)

	err := testCache.Set("testValue", "123")
	assert.NoError(t, err)
}

func TestCache_Get(t *testing.T) {
	db, teardown := repository.TestDB(t)
	defer teardown("\"order\"", "delivery", "payment", "item")

	repo := repository.NewRepository(db)
	testCache := NewCache(repo, 100)

	testCache.Set("testValue", "123")

	data := testCache.Get("testValue")
	assert.Equal(t, data, "123")
}

func TestCache_WarmCache(t *testing.T) {
	db, teardown := repository.TestDB(t)
	defer teardown("\"order\"", "delivery", "payment", "item")

	repo := repository.NewRepository(db)
	testCache := NewCache(repo, 3)

	firstOrder := repository.NewOrder(testDir)
	secondOrder := repository.NewOrder(testDir)
	thirdOrder := repository.NewOrder(testDir)

	repo.Order.Insert(firstOrder)
	repo.Order.Insert(secondOrder)
	repo.Order.Insert(thirdOrder)

	testCache.WarmCache()

	firstOrderCached := testCache.Get(firstOrder.OrderUID.String())
	secondOrderCached := testCache.Get(secondOrder.OrderUID.String())
	thirdOrderCached := testCache.Get(thirdOrder.OrderUID.String())

	assert.Equal(t, firstOrder, firstOrderCached)
	assert.Equal(t, secondOrder, secondOrderCached)
	assert.Equal(t, thirdOrder, thirdOrderCached)
}
