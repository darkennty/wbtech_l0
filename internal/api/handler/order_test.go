package handler

import (
	"WBTech_L0/internal/api/server"
	"WBTech_L0/internal/caches"
	"WBTech_L0/internal/repository"
	"WBTech_L0/internal/service"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetOrderByID(t *testing.T) {
	db, teardown := repository.TestDB(t)
	defer teardown()

	repos := repository.NewRepository(db)
	orderCache := caches.NewCache(repos, 10)
	services := service.NewService(repos)
	handlers := NewHandler(services, orderCache)

	srv := new(server.Server)
	router := handlers.InitRoutes()
	go func() {
		if err := srv.Run("8888", router); err != nil && !errors.Is(http.ErrServerClosed, err) {
			logrus.Fatalf("Error occured while running http-server: %s", err.Error())
		}
	}()

	order := repository.NewOrder("../../testdata/model.json")
	err := repos.Order.Insert(order)
	if err != nil {
		t.Fatal()
	}

	testCases := []struct {
		name         string
		orderUID     string
		expectedCode int
	}{
		{
			name:         "valid",
			orderUID:     order.OrderUID.String(),
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid uid",
			orderUID:     "invalid",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid data",
			orderUID:     uuid.New().String(),
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/order/%s", tc.orderUID), nil)
			router.ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}
