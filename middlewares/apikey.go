package middlewares

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// ApiKeyMiddleware validates the X-API-Key header against EXPORT_API_KEY env var.
func ApiKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		expected := os.Getenv("EXPORT_API_KEY")

		if expected == "" || key != expected {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing API key"})
			c.Abort()
			return
		}
		c.Next()
	}
}
