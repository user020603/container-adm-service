package routes

import (
	"thanhnt208/container-adm-service/api/middlewares"
	"thanhnt208/container-adm-service/internal/delivery/rest"

	"github.com/gin-gonic/gin"
)

func SetupContainerRoutes(h *rest.RestContainerHandler) *gin.Engine {
	router := gin.Default()

	router.POST("/create", h.CreateContainer)
	router.GET("/view", h.ViewContainers)
	router.PUT("/update/:id", middlewares.JWTAuthMiddleware(), middlewares.AdminOnlyMiddleware(), h.UpdateContainer)
	router.DELETE("/delete/:id", middlewares.JWTAuthMiddleware(), middlewares.AdminOnlyMiddleware(), h.DeleteContainer)
	router.POST("/import", h.ImportContainers)
	router.GET("/export", h.ExportContainers)

	return router
}
