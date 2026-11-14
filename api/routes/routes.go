package routes

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"erp-api/config"
	"erp-api/handlers"
	"erp-api/middleware"
	"erp-api/repositories"
	"erp-api/services"
)

func SetupRoutes(router *gin.Engine, cfg config.Config, db *sql.DB) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "ERP API is running",
		})
	})

	// Supplier dependencies
	supplierRepository := repositories.NewSupplierRepository(db)
	supplierService := services.NewSupplierService(supplierRepository)
	supplierHandler := handlers.NewSupplierHandler(supplierService)

	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.LaravelAPIURL))

	api.GET("/me", handlers.Me)

	// Supplier routes
	suppliers := api.Group("/suppliers")
	{
		suppliers.GET("", supplierHandler.GetAll)
		suppliers.GET("/:id", supplierHandler.GetByID)
		suppliers.POST("", supplierHandler.Create)
		suppliers.PUT("/:id", supplierHandler.Update)
		suppliers.DELETE("/:id", supplierHandler.Delete)
	}
}
