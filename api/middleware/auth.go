package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthUser struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	CompanyID   uint64 `json:"company_id"`
	CompanyName string `json:"company_name"`
}

func AuthMiddleware(laravelAPIURL string) gin.HandlerFunc {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" ||
			!strings.HasPrefix(authHeader, "Bearer ") ||
			strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer ")) == "" {

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthenticated",
			})
			return
		}

		req, err := http.NewRequestWithContext(
			c.Request.Context(),
			http.MethodGet,
			strings.TrimRight(laravelAPIURL, "/")+"/api/user",
			nil,
		)

		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"message": "Authentication service unavailable",
			})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized ||
			resp.StatusCode == http.StatusForbidden {

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthenticated",
			})
			return
		}

		if resp.StatusCode != http.StatusOK {
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
				"message": "Unable to validate authentication",
			})
			return
		}

		var user AuthUser

		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
				"message": "Invalid authentication response",
			})
			return
		}

		if user.CompanyID == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "User is not assigned to a company",
			})
			return
		}

		c.Set("authUser", user)

		c.Next()
	}
}
