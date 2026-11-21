package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"erp-api/config"
	"erp-api/routes"
)

// @title ERP API
// @version 1.0.0
// @description ERP API for master data, purchasing, inventory, and B2B sales orders. Endpoints under /api use the Laravel POS token.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Use format: Bearer <token>
func main() {
	cfg := config.LoadConfig()

	db, err := config.ConnectDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}
	defer db.Close()

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	router.StaticFile("/swagger.yaml", "../docs/api/swagger.yaml")

	routes.SetupRoutes(router, cfg, db)

	log.Println("Inventory API running on port:", cfg.AppPort)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
