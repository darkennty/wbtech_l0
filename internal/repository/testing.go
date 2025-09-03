package repository

import (
	"WBTech_L0/internal/model"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestDB(t *testing.T) (*sqlx.DB, func(...string)) {
	t.Helper()

	cfg, err := GetConfig()
	if err != nil {
		t.Fatal(err)
	}

	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode))
	if err != nil {
		t.Fatal()
	}

	if err = db.Ping(); err != nil {
		t.Fatal()
	}

	return db, func(tables ...string) {
		if len(tables) > 0 {
			db.Exec(fmt.Sprintf("TRUNCATE %s CASCADE", strings.Join(tables, ", ")))
		}
		db.Close()
	}
}

func NewOrder(dir string) model.Order {
	message, _ := os.ReadFile(dir)

	var data model.Order
	if err := json.Unmarshal(message, &data); err != nil {
		logrus.Error(fmt.Sprintf("Error while creating new order: %s", err.Error()))
	}

	data.OrderUID = uuid.New()

	return data
}

func GetConfig() (Config, error) {
	dir, err := getProjectRoot()
	if err != nil {
		return Config{}, err
	}

	viper.AddConfigPath(fmt.Sprintf("%s\\configs", dir))
	viper.SetConfigName("config")
	viper.ReadInConfig()

	godotenv.Load(fmt.Sprintf("%s\\.env", dir))

	return Config{
		Host:     viper.GetString("db.host"),
		Port:     os.Getenv("POSTGRES_PORT"),
		Username: os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASS"),
		DBName:   "orders_db_test",
		SSLMode:  viper.GetString("db.ssl_mode"),
	}, nil
}

func getProjectRoot() (string, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error executing 'go list -m': %v, stderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}
