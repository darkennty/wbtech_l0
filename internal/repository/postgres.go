package repository

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	orderTable    = "\"order\""
	deliveryTable = "delivery"
	paymentTable  = "payment"
	itemTable     = "item"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresDB(cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)                 // Максимум открытых соединений
	db.SetMaxIdleConns(10)                 // Максимум idle соединений
	db.SetConnMaxLifetime(5 * time.Minute) // Время жизни соединения
	db.SetConnMaxIdleTime(1 * time.Minute) // Время idle

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
