package kafka_handler

import (
	"WBTech_L0/internal/caches"
	"WBTech_L0/internal/model"
	"WBTech_L0/internal/service"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	service *service.Service
	cache   *caches.Cache
}

func NewHandler(service *service.Service, cache *caches.Cache) *Handler {
	return &Handler{
		service: service,
		cache:   cache,
	}
}

func (h *Handler) HandleMessage(message []byte, topic kafka.TopicPartition, consumerNumber int) error {
	logrus.Infof("Consumer #%d. Message from kafka with offset %d: '%s' on partition %d", consumerNumber, topic.Offset, string(message), topic.Partition)

	var data model.Order
	err := json.Unmarshal(message, &data)
	if err != nil {
		logrus.Error("Error while unmarshaling kafka message into model.Order struct")
		return err
	}

	if err = h.service.Insert(data); err != nil {
		logrus.Error("Error while inserting data into db")
		return err
	}

	if err = h.cache.Set(data.OrderUID.String(), data); err != nil {
		logrus.Error("Error while caching data")
		return err
	}

	return nil
}
