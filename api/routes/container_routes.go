package routes

import (
	"thanhnt208/container-adm-service/api/middlewares"
	"thanhnt208/container-adm-service/internal/delivery/rest"

	"github.com/gin-gonic/gin"
)

func SetupContainerRoutes(h *rest.RestContainerHandler) *gin.Engine {
	router := gin.Default()

	router.POST("/container/create", h.CreateContainer)
	router.GET("/container/view", h.ViewContainers)
	router.PUT("/container/update/:id", middlewares.JWTAuthMiddleware(), middlewares.AdminOnlyMiddleware(), h.UpdateContainer)
	router.DELETE("/container/delete/:id", middlewares.JWTAuthMiddleware(), middlewares.AdminOnlyMiddleware(), h.DeleteContainer)
	router.POST("/container/import", h.ImportContainers)
	router.GET("/container/export", h.ExportContainers)

	return router
}
