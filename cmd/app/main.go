package main

import (
	"WBTech_L0/internal/app"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// @title Orders Service
// @version 1.0
// @description Go-microservice consisting of Kafka consumer and JSON API handler

// @host localhost:8000
// @BasePath /

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Print("Initializing configs...")
	if err := initConfig(); err != nil {
		logrus.Fatalf("Error initializing configs: %s", err.Error())
	}

	logrus.Print("Loading .env variables...")
	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("Error loading .env variables: %s", err.Error())
	}

	app.Run()
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
