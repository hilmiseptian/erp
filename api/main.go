package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"erp-api/config"
	"erp-api/middleware"
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
	configureLogger()
	if err := run(); err != nil {
		slog.Error("ERP API stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.LoadConfig()

	db, err := config.ConnectDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	router := gin.New()
	router.Use(middleware.SlogRequestLogger(), gin.Recovery())

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://localhost:5174",
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

	slog.Info("ERP API listening", "port", cfg.AppPort)

	return router.Run(":" + cfg.AppPort)
}

func configureLogger() {
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler).With("service", "erp-api"))
}
