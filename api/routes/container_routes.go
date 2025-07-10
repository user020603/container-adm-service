package routes

import (
	"thanhnt208/container-adm-service/api/middlewares"
	"thanhnt208/container-adm-service/internal/delivery/rest"

	"github.com/gin-gonic/gin"
)

func SetupContainerRoutes(h rest.IContainerHandler) *gin.Engine {
	router := gin.Default()

	router.POST("/create",
		middlewares.JWTAuthMiddleware(),
		middlewares.CheckScopeMiddleware("container:create"),
		h.CreateContainer,
	)

	router.POST("/view",
		middlewares.JWTAuthMiddleware(),
		middlewares.CheckScopeMiddleware("container:read"),
		h.ViewContainers,
	)

	router.PUT("/update/:id",
		middlewares.JWTAuthMiddleware(),
		middlewares.CheckScopeMiddleware("container:update"),
		h.UpdateContainer,
	)

	router.DELETE("/delete/:id",
		middlewares.JWTAuthMiddleware(),
		middlewares.CheckScopeMiddleware("container:delete"),
		h.DeleteContainer,
	)

	router.POST("/import",
		middlewares.JWTAuthMiddleware(),
		middlewares.CheckScopeMiddleware("container:import"),
		h.ImportContainers,
	)

	router.POST("/export",
		middlewares.JWTAuthMiddleware(),
		middlewares.CheckScopeMiddleware("container:export"),
		h.ExportContainers,
	)

	return router
}
