package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// @Summary      Get Order By ID
// @Description  Handler for getting order from db/cache by its order_uid
// @Tags         orders
// @ID           get-order-by-id
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Order UID (UUID string)"
// @Success      200  {object}  model.Order
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /order/{id} [get]
func (h *Handler) getOrderByID(ctx *gin.Context) {
	orderUID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		NewErrorResponse(ctx, http.StatusBadRequest, "Error while parsing order_uid")
		return
	}

	start := time.Now()
	cachedOrder := h.cache.Get(orderUID.String())
	queryTime := time.Since(start)

	if cachedOrder == nil {
		start = time.Now()
		order, err := h.services.Order.GetOrderByID(orderUID)
		queryTime = time.Since(start)

		if err != nil {
			NewErrorResponse(ctx, http.StatusNotFound, "No orders with such OrderUID")
			return
		}

		if err = h.cache.Set(orderUID.String(), order); err != nil {
			NewErrorResponse(ctx, http.StatusInternalServerError, "Error while caching data")
			return
		}

		logrus.WithFields(logrus.Fields{
			"order_uid": orderUID,
			"time_ms":   queryTime.Milliseconds(),
			"source":    "database",
		}).Info("Order retrieved")
		ctx.JSON(http.StatusOK, order)
	} else {
		logrus.WithFields(logrus.Fields{
			"order_uid": orderUID,
			"time_ms":   queryTime.Milliseconds(),
			"source":    "cache",
		}).Info("Order retrieved")
		ctx.JSON(http.StatusOK, cachedOrder)
	}
}
