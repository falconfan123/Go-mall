#!/usr/bin/env bash

set -euo pipefail

DB_DSN="${DB_DSN:-postgres://root:fht3825099@localhost:5432/mall?sslmode=disable}"

psql "$DB_DSN" <<'SQL'
\set ON_ERROR_STOP on

WITH missing AS (
  SELECT
    p.id AS product_id,
    GREATEST(COALESCE(p.stock, 0), 0) AS total
  FROM products p
  LEFT JOIN inventory i ON i.product_id = p.id
  WHERE i.product_id IS NULL
)
INSERT INTO inventory (product_id, total, sold)
SELECT product_id, total, 0
FROM missing
ORDER BY product_id;

SELECT
  (SELECT count(*) FROM products) AS product_count,
  (SELECT count(*) FROM inventory) AS inventory_count,
  (SELECT count(*)
   FROM products p
   LEFT JOIN inventory i ON i.product_id = p.id
   WHERE i.product_id IS NULL) AS missing_inventory_count;
SQL
