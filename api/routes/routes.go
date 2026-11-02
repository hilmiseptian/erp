package routes

import (
    "github.com/gin-gonic/gin"

    "erp-api/config"
    "erp-api/handlers"
    "erp-api/middleware"
)

func SetupRoutes(router *gin.Engine, cfg config.Config) {
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "success",
            "message": "Inventory API is running",
        })
    })

    api := router.Group("/api")
    api.Use(middleware.AuthMiddleware(cfg.LaravelAPIURL))

    api.GET("/me", handlers.Me)
}