-- name: CreateCategory :exec
INSERT INTO categories (id, name, slug)
VALUES (UNHEX(REPLACE(?, '-', '')), ?, ?);

-- name: CreateProduct :exec
INSERT INTO products (
  id, category_id, sku, name, slug, description, image_url, price_cents, currency, is_active
) VALUES (
  UNHEX(REPLACE(?, '-', '')),
  CASE WHEN ? IS NULL OR ? = '' THEN NULL ELSE UNHEX(REPLACE(?, '-', '')) END,
  ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetProductByID :one
SELECT
  LOWER(CONCAT(
    SUBSTR(HEX(p.id), 1, 8), '-',
    SUBSTR(HEX(p.id), 9, 4), '-',
    SUBSTR(HEX(p.id), 13, 4), '-',
    SUBSTR(HEX(p.id), 17, 4), '-',
    SUBSTR(HEX(p.id), 21)
  )) AS id,
  CASE WHEN p.category_id IS NULL THEN NULL ELSE LOWER(CONCAT(
    SUBSTR(HEX(p.category_id), 1, 8), '-',
    SUBSTR(HEX(p.category_id), 9, 4), '-',
    SUBSTR(HEX(p.category_id), 13, 4), '-',
    SUBSTR(HEX(p.category_id), 17, 4), '-',
    SUBSTR(HEX(p.category_id), 21)
  )) END AS category_id,
  p.sku, p.name, p.slug, p.description, p.image_url,
  p.price_cents, p.currency, p.is_active,
  p.created_at, p.updated_at,
  COALESCE(i.qty, 0) AS qty
FROM products p
LEFT JOIN inventory i ON i.product_id = p.id
WHERE p.id = UNHEX(REPLACE(?, '-', ''))
LIMIT 1;

-- name: ListActiveProducts :many
SELECT
  LOWER(CONCAT(
    SUBSTR(HEX(p.id), 1, 8), '-',
    SUBSTR(HEX(p.id), 9, 4), '-',
    SUBSTR(HEX(p.id), 13, 4), '-',
    SUBSTR(HEX(p.id), 17, 4), '-',
    SUBSTR(HEX(p.id), 21)
  )) AS id,
  p.sku, p.name, p.slug, p.image_url,
  p.price_cents, p.currency,
  COALESCE(i.qty, 0) AS qty
FROM products p
LEFT JOIN inventory i ON i.product_id = p.id
WHERE p.is_active = 1
ORDER BY p.created_at DESC
LIMIT ? OFFSET ?;

-- name: UpsertInventory :exec
INSERT INTO inventory (product_id, qty)
VALUES (UNHEX(REPLACE(?, '-', '')), ?)
ON DUPLICATE KEY UPDATE qty = VALUES(qty);