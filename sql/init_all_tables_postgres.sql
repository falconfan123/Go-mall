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
  order_id TEXT NOT NULL,
  status INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  user_id BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX idx_inventory_lock_order_user ON inventory_lock(order_id, user_id);

DROP TABLE IF EXISTS return_lock CASCADE;
CREATE TABLE return_lock (
  id BIGSERIAL PRIMARY KEY,
  order_id TEXT NOT NULL,
  status INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  user_id BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX idx_return_lock_order_user ON return_lock(order_id, user_id);

-- Users table
DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE users (
  user_id BIGSERIAL PRIMARY KEY,
  username VARCHAR(255) DEFAULT NULL,
  email VARCHAR(255) DEFAULT NULL,
  password_hash VARCHAR(512) DEFAULT NULL,
  avatar_url VARCHAR(255) DEFAULT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  user_deleted BOOLEAN NOT NULL DEFAULT FALSE,
  logout_at TIMESTAMP DEFAULT NULL,
  login_at TIMESTAMP DEFAULT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);

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
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP DEFAULT NULL
);
CREATE INDEX idx_user_address_user_id ON user_address(user_id);

-- Audit table
DROP TABLE IF EXISTS audit CASCADE;
CREATE TABLE audit (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  action_type VARCHAR(64) NOT NULL,
  action_desc TEXT,
  old_data JSONB,
  new_data JSONB,
  service_name VARCHAR(64) NOT NULL,
  target_table VARCHAR(64) NOT NULL,
  target_id BIGINT NOT NULL,
  client_ip VARCHAR(45) NOT NULL,
  trace_id VARCHAR(36) NOT NULL UNIQUE,
  span_id VARCHAR(36) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_audit_user_id ON audit(user_id);
CREATE INDEX idx_audit_service_name ON audit(service_name);
CREATE INDEX idx_audit_action_type ON audit(action_type);
CREATE INDEX idx_audit_target ON audit(target_table, target_id);
CREATE INDEX idx_audit_created_at ON audit(created_at);

-- Coupons tables
DROP TABLE IF EXISTS user_coupons CASCADE;
CREATE TABLE user_coupons (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  coupon_id VARCHAR(36) NOT NULL,
  status INTEGER NOT NULL DEFAULT 0,
  order_id VARCHAR(64),
  used_at TIMESTAMP DEFAULT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (user_id, coupon_id),
  UNIQUE (order_id)
);
CREATE INDEX idx_user_coupons_user_id ON user_coupons(user_id);
CREATE INDEX idx_user_coupons_coupon_id ON user_coupons(coupon_id);

DROP TABLE IF EXISTS coupon_usage CASCADE;
CREATE TABLE coupon_usage (
  id BIGSERIAL PRIMARY KEY,
  order_id VARCHAR(64) NOT NULL,
  coupon_id VARCHAR(36) NOT NULL,
  user_id BIGINT NOT NULL,
  coupon_type INTEGER NOT NULL,
  origin_value BIGINT NOT NULL,
  discount_amount BIGINT NOT NULL,
  applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_coupon_usage_order_id ON coupon_usage(order_id);
CREATE INDEX idx_coupon_usage_user_id ON coupon_usage(user_id);

DROP TABLE IF EXISTS coupons CASCADE;
CREATE TABLE coupons (
  id VARCHAR(36) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  type INTEGER NOT NULL,
  value BIGINT NOT NULL,
  min_amount BIGINT NOT NULL DEFAULT 0,
  start_time TIMESTAMP NOT NULL,
  end_time TIMESTAMP NOT NULL,
  status INTEGER NOT NULL DEFAULT 1,
  total_count INTEGER NOT NULL DEFAULT 0,
  remaining_count INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_coupons_status ON coupons(status);
CREATE INDEX idx_coupons_time ON coupons(start_time, end_time);

INSERT INTO coupons (id, name, type, value, min_amount, start_time, end_time, status, total_count, remaining_count) VALUES
('67508ec1ea7111ef86d80242ac120005', '已领取测试券', 1, 10000, 10000, '2025-01-01 00:00:00', '2027-01-01 00:00:00', 1, 100, 99),
('ZK20250214001', '八折测试券', 2, 80, 0, '2025-01-01 00:00:00', '2027-01-01 00:00:00', 1, 100, 100),
('FJ20250214001', '立减测试券', 3, 100, 0, '2025-01-01 00:00:00', '2027-01-01 00:00:00', 1, 100, 100),
('LOCK20250525001', '锁定测试券', 3, 100, 0, '2025-01-01 00:00:00', '2027-01-01 00:00:00', 1, 100, 100),
('RELEASE20250525001', '释放测试券', 3, 100, 0, '2025-01-01 00:00:00', '2027-01-01 00:00:00', 1, 100, 100),
('USE20250525001', '使用测试券', 3, 100, 0, '2025-01-01 00:00:00', '2027-01-01 00:00:00', 1, 100, 100),
('USED20250525001', '已使用测试券', 3, 100, 0, '2025-01-01 00:00:00', '2027-01-01 00:00:00', 1, 100, 100),
('DUP20250525001', '重复使用测试券', 3, 100, 0, '2025-01-01 00:00:00', '2027-01-01 00:00:00', 1, 100, 100),
('679e623cea7111ef86d80242ac120005', '售罄测试券', 1, 100, 0, '2025-01-01 00:00:00', '2027-01-01 00:00:00', 1, 1, 0)
ON CONFLICT (id) DO NOTHING;

INSERT INTO user_coupons (user_id, coupon_id, status) VALUES
(1, '67508ec1ea7111ef86d80242ac120005', 1),
(1, 'LOCK20250525001', 1),
(1, 'RELEASE20250525001', 1),
(1, 'USE20250525001', 1),
(1, 'USED20250525001', 3),
(1, 'DUP20250525001', 1)
ON CONFLICT (user_id, coupon_id) DO NOTHING;

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
  pre_order_id VARCHAR(64) PRIMARY KEY,
  user_id BIGINT NOT NULL,
  address_id BIGINT NOT NULL,
  coupon_id VARCHAR(255) DEFAULT NULL,
  original_amount BIGINT NOT NULL,
  final_amount BIGINT NOT NULL,
  status SMALLINT NOT NULL DEFAULT 0,
  expire_time BIGINT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_checkouts_user_id ON checkouts(user_id);
CREATE INDEX idx_checkouts_user_status ON checkouts(user_id, status);
CREATE INDEX idx_checkouts_expire_time ON checkouts(expire_time);

DROP TABLE IF EXISTS checkout_items CASCADE;
CREATE TABLE checkout_items (
  id BIGSERIAL PRIMARY KEY,
  pre_order_id VARCHAR(64) NOT NULL,
  product_id BIGINT NOT NULL,
  quantity INTEGER NOT NULL,
  price BIGINT NOT NULL,
  snapshot JSONB NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_checkout_items_pre_order_id ON checkout_items(pre_order_id);
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

-- RAG stats event table
DROP TABLE IF EXISTS rag_stat_events CASCADE;
CREATE TABLE rag_stat_events (
  id BIGSERIAL PRIMARY KEY,
  conversation_id VARCHAR(128) NOT NULL,
  turn_id VARCHAR(128) NOT NULL,
  user_id VARCHAR(128) NOT NULL DEFAULT '',
  app_id VARCHAR(128) NOT NULL,
  knowledge_base_id VARCHAR(128) NOT NULL,
  channel VARCHAR(64) NOT NULL,
  event_type VARCHAR(64) NOT NULL,
  is_rag BOOLEAN NOT NULL DEFAULT FALSE,
  retrieved_doc_count INTEGER NOT NULL DEFAULT 0,
  rag_strategy VARCHAR(128) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT '',
  error_code VARCHAR(128) NOT NULL DEFAULT '',
  latency_ms BIGINT NOT NULL DEFAULT 0,
  trace_id VARCHAR(128) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_rag_stat_conversation_created ON rag_stat_events(conversation_id)
  WHERE event_type = 'conversation_created';
CREATE INDEX idx_rag_stat_created_at ON rag_stat_events(created_at);
CREATE INDEX idx_rag_stat_app_created_at ON rag_stat_events(app_id, created_at);
CREATE INDEX idx_rag_stat_kb_created_at ON rag_stat_events(knowledge_base_id, created_at);
CREATE INDEX idx_rag_stat_channel_created_at ON rag_stat_events(channel, created_at);
CREATE INDEX idx_rag_stat_event_type_created_at ON rag_stat_events(event_type, created_at);
CREATE INDEX idx_rag_stat_turn_id ON rag_stat_events(turn_id);

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
(3, 1), (3, 2), (3, 5);
