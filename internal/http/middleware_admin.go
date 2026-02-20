package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func adminMiddleware(adminToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := c.GetHeader("X-Admin-Token")
		if tok == "" || tok != adminToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
