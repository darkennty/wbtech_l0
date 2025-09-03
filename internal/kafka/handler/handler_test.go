package kafka_handler

import (
	"WBTech_L0/internal/caches"
	"WBTech_L0/internal/repository"
	"WBTech_L0/internal/service"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestHandleMessage(t *testing.T) {
	message, _ := os.ReadFile("../../testdata/model.json")

	db, teardown := repository.TestDB(t)
	defer teardown("\"order\"", "delivery", "payment", "item")

	repo := repository.NewRepository(db)
	cache := caches.NewCache(repo, 5)
	services := service.NewService(repo)
	kafkahandlers := NewHandler(services, cache)

	topic := "testTopic"
	err := kafkahandlers.HandleMessage(message, kafka.TopicPartition{Topic: &topic}, 1)
	assert.NoError(t, err)
}
