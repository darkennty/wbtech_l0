package handler

import (
	_ "WBTech_L0/docs"
	"WBTech_L0/internal/caches"
	"WBTech_L0/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"       // swagger embed files
	"github.com/swaggo/gin-swagger" // gin-swagger middleware
)

type Handler struct {
	services *service.Service
	cache    *caches.Cache
}

func NewHandler(services *service.Service, cache *caches.Cache) *Handler {
	return &Handler{
		services: services,
		cache:    cache,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	order := router.Group("/order")
	{
		order.GET("/:id", h.getOrderByID)
	}

	return router
}
