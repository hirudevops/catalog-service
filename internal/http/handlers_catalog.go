package httpserver

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hirudevops/catalog-service/internal/cache"
	"github.com/hirudevops/catalog-service/internal/config"
	mysqlstore "github.com/hirudevops/catalog-service/internal/store/mysql"
)

type catalogHandlers struct {
	cfg   config.Config
	store *mysqlstore.Store
	cache *cache.Cache
}

func newCatalogHandlers(cfg config.Config, s *mysqlstore.Store, c *cache.Cache) *catalogHandlers {
	return &catalogHandlers{cfg: cfg, store: s, cache: c}
}

func (h *catalogHandlers) Healthz(c *gin.Context) {
	if err := h.store.Health(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "down"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type createCategoryReq struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *catalogHandlers) CreateCategory(c *gin.Context) {
	var req createCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" || req.Slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	err := h.store.CreateCategory(c.Request.Context(), mysqlstore.Category{
		ID:   uuid.New(),
		Name: req.Name,
		Slug: req.Slug,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "created"})
}

type createProductReq struct {
	CategoryID  *string `json:"category_id"` // optional UUID string
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	ImageURL    *string `json:"image_url"`
	PriceCents  int64   `json:"price_cents"`
	Currency    string  `json:"currency"`
	IsActive    *bool   `json:"is_active"`
}

func (h *catalogHandlers) CreateProduct(c *gin.Context) {
	var req createProductReq
	if err := c.ShouldBindJSON(&req); err != nil || req.SKU == "" || req.Name == "" || req.Slug == "" || req.PriceCents <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if req.Currency == "" {
		req.Currency = "BDT"
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}

	var catID *uuid.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		u, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		catID = &u
	}

	var desc sql.NullString
	if req.Description != nil {
		desc = sql.NullString{String: *req.Description, Valid: true}
	}
	var img sql.NullString
	if req.ImageURL != nil {
		img = sql.NullString{String: *req.ImageURL, Valid: true}
	}

	id := uuid.New()
	err := h.store.CreateProduct(c.Request.Context(), mysqlstore.Product{
		ID:          id,
		CategoryID:  catID,
		SKU:         req.SKU,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: desc,
		ImageURL:    img,
		PriceCents:  req.PriceCents,
		Currency:    req.Currency,
		IsActive:    active,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Initialize inventory = 0
	_ = h.store.UpsertInventory(c.Request.Context(), id, 0)

	c.JSON(http.StatusCreated, gin.H{"id": id.String()})
}

func (h *catalogHandlers) UpsertInventory(c *gin.Context) {
	pid, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product_id"})
		return
	}

	type reqBody struct {
		Qty int64 `json:"qty"`
	}
	var req reqBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if err := h.store.UpsertInventory(c.Request.Context(), pid, req.Qty); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Bust product cache
	_ = h.cache.Rdb.Del(c.Request.Context(), fmt.Sprintf("cache:catalog:product:%s", pid.String())).Err()

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type productResp struct {
	ID         string  `json:"id"`
	CategoryID *string `json:"category_id,omitempty"`
	SKU        string  `json:"sku"`
	Name       string  `json:"name"`
	Slug       string  `json:"slug"`
	ImageURL   *string `json:"image_url,omitempty"`
	PriceCents int64   `json:"price_cents"`
	Currency   string  `json:"currency"`
	Qty        int64   `json:"qty"`
}

func (h *catalogHandlers) GetProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	cacheKey := fmt.Sprintf("cache:catalog:product:%s", id.String())

	var cached productResp
	if ok, err := h.cache.GetJSON(c.Request.Context(), cacheKey, &cached); err == nil && ok {
		c.JSON(http.StatusOK, cached)
		return
	}

	p, err := h.store.GetProductByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	resp := productResp{
		ID:         p.ID.String(),
		SKU:        p.SKU,
		Name:       p.Name,
		Slug:       p.Slug,
		PriceCents: p.PriceCents,
		Currency:   p.Currency,
		Qty:        p.Qty,
	}
	if p.CategoryID != nil {
		s := p.CategoryID.String()
		resp.CategoryID = &s
	}
	if p.ImageURL.Valid {
		resp.ImageURL = &p.ImageURL.String
	}

	_ = h.cache.SetJSON(c.Request.Context(), cacheKey, resp, h.cfg.CacheProductTTL)
	c.JSON(http.StatusOK, resp)
}

func (h *catalogHandlers) ListProducts(c *gin.Context) {
	limit := parseIntDefault(c.Query("limit"), 20)
	offset := parseIntDefault(c.Query("offset"), 0)
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	cacheKey := fmt.Sprintf("cache:catalog:product_list:l:%d:o:%d", limit, offset)
	var cached []productResp
	if ok, err := h.cache.GetJSON(c.Request.Context(), cacheKey, &cached); err == nil && ok {
		c.JSON(http.StatusOK, gin.H{"items": cached})
		return
	}

	items, err := h.store.ListActiveProducts(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := make([]productResp, 0, len(items))
	for _, p := range items {
		r := productResp{
			ID:         p.ID.String(),
			SKU:        p.SKU,
			Name:       p.Name,
			Slug:       p.Slug,
			PriceCents: p.PriceCents,
			Currency:   p.Currency,
			Qty:        p.Qty,
		}
		if p.ImageURL.Valid {
			r.ImageURL = &p.ImageURL.String
		}
		resp = append(resp, r)
	}

	_ = h.cache.SetJSON(c.Request.Context(), cacheKey, resp, h.cfg.CacheProductListTTL)
	c.JSON(http.StatusOK, gin.H{"items": resp})
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
