package main

import (
    "log"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"

    "erp-api/config"
    "erp-api/routes"
)

func main() {
    cfg := config.LoadConfig()

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
        MaxAge: 12 * time.Hour,
    }))

    routes.SetupRoutes(router, cfg)

    log.Println("Inventory API running on port:", cfg.AppPort)

    if err := router.Run(":" + cfg.AppPort); err != nil {
        log.Fatal(err)
    }
}