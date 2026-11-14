package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "erp-api/middleware"
)

// Me returns the authenticated POS user profile.
// @Summary Current user
// @Tags Auth
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/me [get]
func Me(c *gin.Context) {
    user, exists := c.Get("authUser")

    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{
            "message": "Unauthenticated",
        })
        return
    }

    authUser := user.(middleware.AuthUser)

    c.JSON(http.StatusOK, gin.H{
        "message": "Authenticated successfully",
        "user":    authUser,
    })
}
