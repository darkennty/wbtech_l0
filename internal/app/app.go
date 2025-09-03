package app

import (
	"WBTech_L0/internal/api/handler"
	"WBTech_L0/internal/api/server"
	"WBTech_L0/internal/caches"
	"WBTech_L0/internal/kafka/consumer"
	kafkahandler "WBTech_L0/internal/kafka/handler"
	"WBTech_L0/internal/repository"
	"WBTech_L0/internal/service"
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	capacity = 100
)

func Run() {
	logrus.Print("Initializing DB...")
	db, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     os.Getenv("POSTGRES_PORT"),
		Username: os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASS"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  viper.GetString("db.ssl_mode"),
	})
	if err != nil {
		logrus.Fatalf("Error initializing DB: %s", err.Error())
	}

	logrus.Print("Initializing components...")
	repos := repository.NewRepository(db)
	orderCache := caches.NewCache(repos, capacity)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services, orderCache)

	if err = orderCache.WarmCache(); err != nil {
		logrus.Fatalf("Error warming cache: %s", err.Error())
	}

	srv := new(server.Server)
	go func() {
		if err = srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil && !errors.Is(http.ErrServerClosed, err) {
			logrus.Fatalf("Error occured while running http-server: %s", err.Error())
		}
	}()

	logrus.Print("Setting up kafka...")
	h := kafkahandler.NewHandler(services, orderCache)
	addresses := []string{viper.GetString("kafka.address_1"), viper.GetString("kafka.address_2"), viper.GetString("kafka.address_3")}
	topic := viper.GetString("kafka.topic")
	consumerGroup := viper.GetString("kafka.consumer_group")

	c1, err := consumer.NewConsumer(h, addresses, topic, consumerGroup, 1)
	if err != nil {
		logrus.Fatal(err)
	}
	c2, err := consumer.NewConsumer(h, addresses, topic, consumerGroup, 2)
	if err != nil {
		logrus.Fatal(err)
	}
	c3, err := consumer.NewConsumer(h, addresses, topic, consumerGroup, 3)
	if err != nil {
		logrus.Fatal(err)
	}

	go func() {
		c1.Start()
	}()
	go func() {
		c2.Start()
	}()
	go func() {
		c3.Start()
	}()

	logrus.Print("App started.")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("App is shutting down.")

	if err = c1.Stop(); err != nil {
		logrus.Fatal(err)
	}
	if err = c2.Stop(); err != nil {
		logrus.Fatal(err)
	}
	if err = c3.Stop(); err != nil {
		logrus.Fatal(err)
	}

	logrus.Print("Consumers are stopped.")

	if err = srv.Shutdown(context.Background()); err != nil {
		logrus.Fatalf("Error occured while shutting down: %s", err.Error())
	}

	if err = db.Close(); err != nil {
		logrus.Fatalf("Error occured while closing DB: %s", err.Error())
	}

	logrus.Print("App is stopped.")
}
