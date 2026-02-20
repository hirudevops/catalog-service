package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/hirudevops/catalog-service/internal/cache"
	"github.com/hirudevops/catalog-service/internal/config"
	mysqlstore "github.com/hirudevops/catalog-service/internal/store/mysql"
)

type Server struct {
	engine *gin.Engine
}

func New(cfg config.Config) (*Server, error) {
	store, err := mysqlstore.New(cfg.MySQLDSN)
	if err != nil {
		return nil, err
	}

	c := cache.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(requestIDMiddleware())
	r.Use(corsMiddleware(cfg))

	h := newCatalogHandlers(cfg, store, c)

	r.GET("/healthz", h.Healthz)

	catalog := r.Group("/catalog")
	{
		catalog.GET("/products", h.ListProducts)
		catalog.GET("/products/:id", h.GetProduct)
	}

	admin := r.Group("/catalog")
	admin.Use(adminMiddleware(cfg.AdminToken))
	{
		admin.POST("/categories", h.CreateCategory)
		admin.POST("/products", h.CreateProduct)
		admin.PUT("/inventory/:product_id", h.UpsertInventory)
	}

	return &Server{engine: r}, nil
}

func (s *Server) Run(addr string) error {
	return s.engine.Run(addr)
}
