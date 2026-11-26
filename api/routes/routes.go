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
	customerRepository := repositories.NewCustomerRepository(db)
	customerService := services.NewCustomerService(customerRepository)
	customerHandler := handlers.NewCustomerHandler(customerService)
	customers := api.Group("/customers")
	{
		customers.GET("", customerHandler.List)
		customers.GET("/:id", customerHandler.Get)
		customers.POST("", customerHandler.Create)
		customers.PUT("/:id", customerHandler.Update)
		customers.PATCH("/:id/status", customerHandler.SetActive)
	}

	salesOrderRepository := repositories.NewSalesOrderRepository(db)
	salesOrderService := services.NewSalesOrderService(salesOrderRepository)
	salesOrderHandler := handlers.NewSalesOrderHandler(salesOrderService)
	salesOrders := api.Group("/sales-orders")
	{
		salesOrders.GET("/branches", salesOrderHandler.Branches)
		salesOrders.GET("", salesOrderHandler.List)
		salesOrders.GET("/:id", salesOrderHandler.Get)
		salesOrders.POST("", salesOrderHandler.Create)
		salesOrders.PUT("/:id", salesOrderHandler.Update)
		salesOrders.PATCH("/:id/cancel", salesOrderHandler.Cancel)
	}

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

	poRepository := repositories.NewPurchaseOrderRepository(db)
	poService := services.NewPurchaseOrderService(poRepository)
	poHandler := handlers.NewPurchaseOrderHandler(poService)
	purchaseOrders := api.Group("/purchase-orders")
	{
		purchaseOrders.GET("/purchase-requests", poHandler.Sources)
		purchaseOrders.GET("/purchase-requests/:pr_number", poHandler.Source)
		purchaseOrders.GET("", poHandler.List)
		purchaseOrders.GET("/:id", poHandler.Get)
		purchaseOrders.POST("", poHandler.Create)
		purchaseOrders.PATCH("/:id/cancel", poHandler.Cancel)
	}

	grRepository := repositories.NewGoodReceiptRepository(db)
	grService := services.NewGoodReceiptService(grRepository)
	grHandler := handlers.NewGoodReceiptHandler(grService)
	goodReceipts := api.Group("/good-receipts")
	{
		goodReceipts.GET("/purchase-orders", grHandler.Sources)
		goodReceipts.GET("/purchase-orders/:po_number", grHandler.Source)
		goodReceipts.GET("", grHandler.List)
		goodReceipts.GET("/:id", grHandler.Get)
		goodReceipts.POST("", grHandler.Create)
	}

	gdRepository := repositories.NewGoodDeliveryRepository(db)
	gdService := services.NewGoodDeliveryService(gdRepository)
	gdHandler := handlers.NewGoodDeliveryHandler(gdService)
	goodDeliveries := api.Group("/good-deliveries")
	{
		goodDeliveries.GET("/sources", gdHandler.Sources)
		goodDeliveries.GET("/sources/:ref_type/:ref_num", gdHandler.Source)
		goodDeliveries.GET("", gdHandler.List)
		goodDeliveries.GET("/:id", gdHandler.Get)
		goodDeliveries.POST("", gdHandler.Create)
	}

	stockRepository := repositories.NewStockRepository(db)
	stockService := services.NewStockService(stockRepository)
	stockHandler := handlers.NewStockHandler(stockService)
	stocks := api.Group("/stocks")
	{
		stocks.GET("/branches", stockHandler.Branches)
		stocks.GET("", stockHandler.List)
		stocks.GET("/:branch_id/:item_id/cards", stockHandler.Cards)
	}

	stockAdjustmentRepository := repositories.NewStockAdjustmentRepository(db)
	stockAdjustmentService := services.NewStockAdjustmentService(stockAdjustmentRepository)
	stockAdjustmentHandler := handlers.NewStockAdjustmentHandler(stockAdjustmentService)
	stockAdjustments := api.Group("/stock-adjustments")
	{
		stockAdjustments.GET("", stockAdjustmentHandler.List)
		stockAdjustments.GET("/:id", stockAdjustmentHandler.Get)
		stockAdjustments.POST("", stockAdjustmentHandler.Create)
		stockAdjustments.PUT("/:id", stockAdjustmentHandler.Update)
		stockAdjustments.POST("/:id/post", stockAdjustmentHandler.Post)
		stockAdjustments.PATCH("/:id/cancel", stockAdjustmentHandler.Cancel)
	}

	stockTransferRepository := repositories.NewStockTransferRepository(db)
	stockTransferService := services.NewStockTransferService(stockTransferRepository)
	stockTransferHandler := handlers.NewStockTransferHandler(stockTransferService)
	stockTransfers := api.Group("/stock-transfers")
	{
		stockTransfers.GET("/branches", stockTransferHandler.Branches)
		stockTransfers.GET("/items", stockTransferHandler.Items)
		stockTransfers.GET("", stockTransferHandler.List)
		stockTransfers.GET("/:id", stockTransferHandler.Get)
		stockTransfers.POST("", stockTransferHandler.Create)
		stockTransfers.PUT("/:id", stockTransferHandler.Update)
		stockTransfers.POST("/:id/dispatches", stockTransferHandler.Dispatch)
		stockTransfers.POST("/:id/receipts", stockTransferHandler.Receive)
		stockTransfers.PATCH("/:id/cancel", stockTransferHandler.Cancel)
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
