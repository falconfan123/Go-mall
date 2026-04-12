-- PostgreSQL Database Initialization Script
-- Converted from MySQL to PostgreSQL

-- Create database if not exists (run separately)
-- CREATE DATABASE mall;

-- Set schema
SET search_path TO public;

-- Inventory table
DROP TABLE IF EXISTS inventory CASCADE;
CREATE TABLE inventory (
  product_id BIGINT NOT NULL,
  total BIGINT NOT NULL DEFAULT 0,
  sold BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (product_id)
);

-- Inventory lock tables
DROP TABLE IF EXISTS inventory_lock CASCADE;
CREATE TABLE inventory_lock (
  id BIGSERIAL PRIMARY KEY,
  product_id BIGINT NOT NULL,
  quantity INTEGER NOT NULL,
  order_id BIGINT NOT NULL,
  status INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_inventory_lock_product_id ON inventory_lock(product_id);

DROP TABLE IF EXISTS return_lock CASCADE;
CREATE TABLE return_lock (
  id BIGSERIAL PRIMARY KEY,
  product_id BIGINT NOT NULL,
  quantity INTEGER NOT NULL,
  order_id BIGINT NOT NULL,
  status INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_return_lock_product_id ON return_lock(product_id);

-- User address table
DROP TABLE IF EXISTS user_address CASCADE;
CREATE TABLE user_address (
  address_id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  recipient_name VARCHAR(255) NOT NULL,
  phone_number VARCHAR(20) NOT NULL,
  province VARCHAR(255) NOT NULL,
  city VARCHAR(255) NOT NULL,
  detailed_address TEXT NOT NULL,
  is_default INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP DEFAULT NULL
);
CREATE INDEX idx_user_address_user_id ON user_address(user_id);

-- Audit table
DROP TABLE IF EXISTS audit CASCADE;
CREATE TABLE audit (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT,
  action VARCHAR(255) NOT NULL,
  resource VARCHAR(255),
  resource_id BIGINT,
  ip_address VARCHAR(50),
  user_agent TEXT,
  request_data JSONB,
  response_data JSONB,
  status VARCHAR(50),
  error_message TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_audit_user_id ON audit(user_id);
CREATE INDEX idx_audit_action ON audit(action);
CREATE INDEX idx_audit_created_at ON audit(created_at);

-- Coupons tables
DROP TABLE IF EXISTS user_coupons CASCADE;
CREATE TABLE user_coupons (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  coupon_id BIGINT NOT NULL,
  status INTEGER NOT NULL DEFAULT 0,
  order_id BIGINT,
  used_at TIMESTAMP DEFAULT NULL,
  expires_at TIMESTAMP DEFAULT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_user_coupons_user_id ON user_coupons(user_id);
CREATE INDEX idx_user_coupons_coupon_id ON user_coupons(coupon_id);

DROP TABLE IF EXISTS coupon_usage CASCADE;
CREATE TABLE coupon_usage (
  id BIGSERIAL PRIMARY KEY,
  user_coupon_id BIGINT NOT NULL,
  order_id BIGINT NOT NULL,
  used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_coupon_usage_user_coupon_id ON coupon_usage(user_coupon_id);
CREATE INDEX idx_coupon_usage_order_id ON coupon_usage(order_id);

DROP TABLE IF EXISTS coupons CASCADE;
CREATE TABLE coupons (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  type INTEGER NOT NULL,
  discount_value DECIMAL(10, 2) NOT NULL,
  min_order_amount DECIMAL(10, 2) DEFAULT 0,
  total_quantity INTEGER NOT NULL DEFAULT 0,
  used_quantity INTEGER NOT NULL DEFAULT 0,
  per_user_limit INTEGER DEFAULT 1,
  starts_at TIMESTAMP DEFAULT NULL,
  ends_at TIMESTAMP DEFAULT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP DEFAULT NULL
);

-- Cart table
DROP TABLE IF EXISTS carts CASCADE;
CREATE TABLE carts (
  id SERIAL PRIMARY KEY,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP DEFAULT NULL,
  user_id INTEGER,
  product_id INTEGER,
  product_name VARCHAR(255),
  product_image VARCHAR(512),
  product_price DECIMAL(10,2),
  quantity INTEGER,
  checked INTEGER
);
CREATE INDEX idx_carts_user_id ON carts(user_id);
CREATE INDEX idx_carts_product_id ON carts(product_id);

-- Checkout tables
DROP TABLE IF EXISTS checkouts CASCADE;
CREATE TABLE checkouts (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  address_id BIGINT NOT NULL,
  coupon_id BIGINT,
  status INTEGER NOT NULL DEFAULT 0,
  total_amount DECIMAL(10, 2) NOT NULL,
  discount_amount DECIMAL(10, 2) DEFAULT 0,
  final_amount DECIMAL(10, 2) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_checkouts_user_id ON checkouts(user_id);

DROP TABLE IF EXISTS checkout_items CASCADE;
CREATE TABLE checkout_items (
  id BIGSERIAL PRIMARY KEY,
  checkout_id BIGINT NOT NULL,
  product_id BIGINT NOT NULL,
  quantity INTEGER NOT NULL,
  price DECIMAL(10, 2) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_checkout_items_checkout_id ON checkout_items(checkout_id);
CREATE INDEX idx_checkout_items_product_id ON checkout_items(product_id);

-- Order tables
DROP TABLE IF EXISTS orders CASCADE;
CREATE TABLE orders (
  order_id VARCHAR(64) NOT NULL,
  pre_order_id VARCHAR(64),
  user_id BIGINT NOT NULL,
  coupon_id VARCHAR(64),
  payment_method INTEGER,
  transaction_id VARCHAR(64),
  paid_at BIGINT,
  original_amount BIGINT NOT NULL,
  discount_amount BIGINT DEFAULT 0,
  payable_amount BIGINT NOT NULL,
  paid_amount BIGINT,
  order_status INTEGER NOT NULL DEFAULT 0,
  payment_status INTEGER NOT NULL DEFAULT 0,
  reason VARCHAR(255),
  expire_time BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (order_id)
);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_order_status ON orders(order_status);
CREATE INDEX idx_orders_payment_status ON orders(payment_status);

DROP TABLE IF EXISTS order_items CASCADE;
CREATE TABLE order_items (
  order_id VARCHAR(64) NOT NULL,
  product_id BIGINT NOT NULL,
  quantity BIGINT NOT NULL,
  price BIGINT NOT NULL,
  product_name VARCHAR(255) NOT NULL,
  product_desc TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (order_id, product_id)
);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);

DROP TABLE IF EXISTS order_addresses CASCADE;
CREATE TABLE order_addresses (
  address_id BIGSERIAL PRIMARY KEY,
  order_id VARCHAR(64) NOT NULL,
  recipient_name VARCHAR(255) NOT NULL,
  phone_number VARCHAR(20),
  province VARCHAR(255),
  city VARCHAR(255) NOT NULL,
  detailed_address TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (order_id)
);

-- Payment table - using the actual schema from paymentsmodel_gen.go
DROP TABLE IF EXISTS payments CASCADE;
CREATE TABLE payments (
  payment_id VARCHAR(64) NOT NULL,
  pre_order_id VARCHAR(64) NOT NULL,
  order_id VARCHAR(64),
  user_id BIGINT NOT NULL,
  original_amount BIGINT NOT NULL,
  paid_amount BIGINT,
  payment_method VARCHAR(50),
  transaction_id VARCHAR(255),
  pay_url TEXT,
  expire_time BIGINT NOT NULL,
  status BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  paid_at BIGINT,
  PRIMARY KEY (payment_id)
);
CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(status);

-- Insert some test inventory data
INSERT INTO inventory (product_id, total, sold) VALUES
(1, 100, 0),
(2, 50, 0),
(3, 200, 0);

-- Categories table
DROP TABLE IF EXISTS product_categories CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS products CASCADE;

-- Products table
CREATE TABLE products (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  picture TEXT,
  price BIGINT NOT NULL DEFAULT 0,
  stock BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_products_name ON products(name);

-- Categories table
CREATE TABLE categories (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  description TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Product-Category relationship table
CREATE TABLE product_categories (
  id BIGSERIAL PRIMARY KEY,
  product_id BIGINT NOT NULL,
  category_id BIGINT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (product_id, category_id)
);
CREATE INDEX idx_product_categories_product_id ON product_categories(product_id);
CREATE INDEX idx_product_categories_category_id ON product_categories(category_id);

-- Insert some test products
INSERT INTO products (name, description, picture, price, stock) VALUES
('iPhone 15 Pro', 'Apple iPhone 15 Pro 256GB', 'https://via.placeholder.com/300', 899900, 100),
('MacBook Pro 14', 'Apple MacBook Pro 14 inch M3', 'https://via.placeholder.com/300', 1999900, 50),
('AirPods Pro', 'Apple AirPods Pro 2nd Generation', 'https://via.placeholder.com/300', 249900, 200);

-- Insert some test categories
INSERT INTO categories (name, description) VALUES
('Electronics', 'Electronic devices and accessories'),
('Apple', 'Apple products'),
('Smartphones', 'Mobile phones'),
('Laptops', 'Laptop computers'),
('Audio', 'Audio equipment');

-- Insert product-category relationships
INSERT INTO product_categories (product_id, category_id) VALUES
(1, 1), (1, 2), (1, 3),
(2, 1), (2, 2), (2, 4),
(3, 1), (2, 2), (3, 5);
