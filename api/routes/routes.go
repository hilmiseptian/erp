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
	itemRepository := repositories.NewItemRepository(db)
	itemService := services.NewItemService(itemRepository)
	itemHandler := handlers.NewItemHandler(itemService)
	items := api.Group("/items")
	{
		items.GET("", itemHandler.List)
		items.GET("/categories", itemHandler.Categories)
		items.GET("/:id", itemHandler.Get)
		items.POST("", itemHandler.Create)
		items.PUT("/:id", itemHandler.Update)
		items.PATCH("/:id/status", itemHandler.SetActive)
	}

	prRepository := repositories.NewPurchaseRequestRepository(db)
	prService := services.NewPurchaseRequestService(prRepository)
	prHandler := handlers.NewPurchaseRequestHandler(prService)
	purchaseRequests := api.Group("/purchase-requests")
	{
		purchaseRequests.GET("/branches", prHandler.Branches)
		purchaseRequests.GET("/suppliers", prHandler.Suppliers)
		purchaseRequests.GET("", prHandler.List)
		purchaseRequests.GET("/:id", prHandler.Get)
		purchaseRequests.POST("", prHandler.Create)
		purchaseRequests.PUT("/:id", prHandler.Update)
		purchaseRequests.PATCH("/:id/cancel", prHandler.Cancel)
	}

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
