package mysqlstore

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Store struct {
	DB *sqlx.DB
}

func New(dsn string) (*Store, error) {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	return &Store{DB: db}, nil
}

type Category struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	Slug      string    `db:"slug"`
	CreatedAt time.Time `db:"created_at"`
}

type Product struct {
	ID          uuid.UUID      `db:"id"`
	CategoryID  *uuid.UUID     `db:"category_id"`
	SKU         string         `db:"sku"`
	Name        string         `db:"name"`
	Slug        string         `db:"slug"`
	Description sql.NullString `db:"description"`
	ImageURL    sql.NullString `db:"image_url"`
	PriceCents  int64          `db:"price_cents"`
	Currency    string         `db:"currency"`
	IsActive    bool           `db:"is_active"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
	Qty         int64          `db:"qty"`
}

func (s *Store) Health(ctx context.Context) error {
	return s.DB.PingContext(ctx)
}

func (s *Store) CreateCategory(ctx context.Context, c Category) error {
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO categories (id, name, slug) VALUES (?, ?, ?)`,
		uuidToBin(c.ID), c.Name, c.Slug,
	)
	return err
}

func (s *Store) CreateProduct(ctx context.Context, p Product) error {
	var cat any = nil
	if p.CategoryID != nil {
		cat = uuidToBin(*p.CategoryID)
	}

	_, err := s.DB.ExecContext(ctx, `
INSERT INTO products (
  id, category_id, sku, name, slug, description, image_url, price_cents, currency, is_active
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		uuidToBin(p.ID),
		cat,
		p.SKU,
		p.Name,
		p.Slug,
		p.Description,
		p.ImageURL,
		p.PriceCents,
		p.Currency,
		p.IsActive,
	)
	return err
}

func (s *Store) UpsertInventory(ctx context.Context, productID uuid.UUID, qty int64) error {
	_, err := s.DB.ExecContext(ctx, `
INSERT INTO inventory (product_id, qty)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE qty = VALUES(qty)`,
		uuidToBin(productID), qty,
	)
	return err
}

func (s *Store) GetProductByID(ctx context.Context, id uuid.UUID) (Product, error) {
	var p Product
	err := s.DB.GetContext(ctx, &p, `
SELECT
  p.id,
  p.category_id,
  p.sku, p.name, p.slug, p.description, p.image_url,
  p.price_cents, p.currency, p.is_active,
  p.created_at, p.updated_at,
  COALESCE(i.qty, 0) AS qty
FROM products p
LEFT JOIN inventory i ON i.product_id = p.id
WHERE p.id = ?
LIMIT 1`, uuidToBin(id))
	return p, err
}

func (s *Store) ListActiveProducts(ctx context.Context, limit, offset int) ([]Product, error) {
	out := make([]Product, 0)
	err := s.DB.SelectContext(ctx, &out, `
SELECT
  p.id,
  p.category_id,
  p.sku, p.name, p.slug, p.description, p.image_url,
  p.price_cents, p.currency, p.is_active,
  p.created_at, p.updated_at,
  COALESCE(i.qty, 0) AS qty
FROM products p
LEFT JOIN inventory i ON i.product_id = p.id
WHERE p.is_active = 1
ORDER BY p.created_at DESC
LIMIT ? OFFSET ?`, limit, offset)
	return out, err
}

// --- helpers

func uuidToBin(id uuid.UUID) []byte {
	b, _ := id.MarshalBinary()
	return b
}
