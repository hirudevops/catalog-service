package httpserver

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hirudevops/catalog-service/internal/config"
)

func corsMiddleware(cfg config.Config) gin.HandlerFunc {
	allowed := cfg.CORSAllowOrigins

	// If you want cookie-based calls later, set AllowCredentials=true.
	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// exact match
			for _, a := range allowed {
				if origin == a {
					return true
				}
			}
			return false
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-Id", "X-Admin-Token"},
		ExposeHeaders:    []string{"X-Request-Id"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
		// Sometimes browsers send "null" origin in dev; keep it false by default
		AllowWildcard: false,
	})
}

var _ = strings.TrimSpace
