USE mall;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- Inventory table
DROP TABLE IF EXISTS `inventory`;
CREATE TABLE `inventory` (
  `product_id` bigint NOT NULL,
  `total` bigint NOT NULL DEFAULT 0,
  `sold` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`product_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

-- Inventory lock tables
DROP TABLE IF EXISTS `inventory_lock`;
CREATE TABLE `inventory_lock` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `product_id` bigint NOT NULL,
  `quantity` int NOT NULL,
  `order_id` bigint NOT NULL,
  `status` tinyint NOT NULL DEFAULT 0,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `idx_product_id` (`product_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

DROP TABLE IF EXISTS `return_lock`;
CREATE TABLE `return_lock` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `product_id` bigint NOT NULL,
  `quantity` int NOT NULL,
  `order_id` bigint NOT NULL,
  `status` tinyint NOT NULL DEFAULT 0,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `idx_product_id` (`product_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

-- User address table
DROP TABLE IF EXISTS `user_address`;
CREATE TABLE `user_address` (
  `address_id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `recipient_name` varchar(255) NOT NULL,
  `phone_number` varchar(20) NOT NULL,
  `province` varchar(255) NOT NULL,
  `city` varchar(255) NOT NULL,
  `detailed_address` text NOT NULL,
  `is_default` tinyint(1) NOT NULL DEFAULT 0,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`address_id`),
  INDEX `idx_user_id` (`user_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

-- Audit table
DROP TABLE IF EXISTS `audit`;
CREATE TABLE `audit` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NULL,
  `action` varchar(255) NOT NULL,
  `resource` varchar(255) NULL,
  `resource_id` bigint NULL,
  `ip_address` varchar(50) NULL,
  `user_agent` text NULL,
  `request_data` json NULL,
  `response_data` json NULL,
  `status` varchar(50) NULL,
  `error_message` text NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_action` (`action`),
  INDEX `idx_created_at` (`created_at`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

-- Coupons tables
DROP TABLE IF EXISTS `user_coupons`;
CREATE TABLE `user_coupons` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `coupon_id` bigint NOT NULL,
  `status` tinyint NOT NULL DEFAULT 0,
  `order_id` bigint NULL,
  `used_at` timestamp NULL DEFAULT NULL,
  `expires_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_coupon_id` (`coupon_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

DROP TABLE IF EXISTS `coupon_usage`;
CREATE TABLE `coupon_usage` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_coupon_id` bigint NOT NULL,
  `order_id` bigint NOT NULL,
  `used_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `idx_user_coupon_id` (`user_coupon_id`),
  INDEX `idx_order_id` (`order_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

DROP TABLE IF EXISTS `coupons`;
CREATE TABLE `coupons` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `description` text NULL,
  `type` tinyint NOT NULL,
  `discount_value` decimal(10, 2) NOT NULL,
  `min_order_amount` decimal(10, 2) NULL DEFAULT 0,
  `total_quantity` int NOT NULL DEFAULT 0,
  `used_quantity` int NOT NULL DEFAULT 0,
  `per_user_limit` int NULL DEFAULT 1,
  `starts_at` timestamp NULL DEFAULT NULL,
  `ends_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

-- Cart table (already exists but let's make sure)
DROP TABLE IF EXISTS `carts`;
CREATE TABLE `carts` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '主键 自增',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  `user_id` int NULL DEFAULT NULL COMMENT '用户ID',
  `product_id` int NULL DEFAULT NULL COMMENT '商品ID',
  `quantity` int NULL DEFAULT NULL COMMENT '商品数量',
  `checked` tinyint(1) NULL DEFAULT NULL COMMENT '商品是否选中',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_carts_user_id`(`user_id` ASC) USING BTREE,
  INDEX `idx_carts_product_id`(`product_id` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- Checkout tables
DROP TABLE IF EXISTS `checkouts`;
CREATE TABLE `checkouts` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `address_id` bigint NOT NULL,
  `coupon_id` bigint NULL,
  `status` tinyint NOT NULL DEFAULT 0,
  `total_amount` decimal(10, 2) NOT NULL,
  `discount_amount` decimal(10, 2) NULL DEFAULT 0,
  `final_amount` decimal(10, 2) NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `idx_user_id` (`user_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

DROP TABLE IF EXISTS `checkout_items`;
CREATE TABLE `checkout_items` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `checkout_id` bigint NOT NULL,
  `product_id` bigint NOT NULL,
  `quantity` int NOT NULL,
  `price` decimal(10, 2) NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `idx_checkout_id` (`checkout_id`),
  INDEX `idx_product_id` (`product_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

-- Order tables
DROP TABLE IF EXISTS `orders`;
CREATE TABLE `orders` (
  `order_id` varchar(64) NOT NULL,
  `pre_order_id` varchar(64) NULL,
  `user_id` bigint NOT NULL,
  `coupon_id` varchar(64) NULL,
  `payment_method` tinyint NULL,
  `transaction_id` varchar(64) NULL,
  `paid_at` bigint NULL,
  `original_amount` bigint NOT NULL,
  `discount_amount` bigint NULL DEFAULT 0,
  `payable_amount` bigint NOT NULL,
  `paid_amount` bigint NULL,
  `order_status` tinyint NOT NULL DEFAULT 0,
  `payment_status` tinyint NOT NULL DEFAULT 0,
  `reason` varchar(255) NULL,
  `expire_time` bigint NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`order_id`),
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_order_status` (`order_status`),
  INDEX `idx_payment_status` (`payment_status`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

DROP TABLE IF EXISTS `order_items`;
CREATE TABLE `order_items` (
  `order_id` varchar(64) NOT NULL,
  `product_id` bigint NOT NULL,
  `quantity` bigint NOT NULL,
  `price` bigint NOT NULL,
  `product_name` varchar(255) NOT NULL,
  `product_desc` text NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`order_id`, `product_id`),
  INDEX `idx_product_id` (`product_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

DROP TABLE IF EXISTS `order_addresses`;
CREATE TABLE `order_addresses` (
  `address_id` bigint NOT NULL AUTO_INCREMENT,
  `order_id` varchar(64) NOT NULL,
  `recipient_name` varchar(255) NOT NULL,
  `phone_number` varchar(20) NULL,
  `province` varchar(255) NULL,
  `city` varchar(255) NOT NULL,
  `detailed_address` text NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`address_id`),
  UNIQUE INDEX `idx_order_id` (`order_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

-- Payment table
DROP TABLE IF EXISTS `payments`;
CREATE TABLE `payments` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `order_id` bigint NOT NULL,
  `payment_no` varchar(64) NOT NULL,
  `amount` decimal(10, 2) NOT NULL,
  `status` tinyint NOT NULL DEFAULT 0,
  `payment_method` varchar(50) NULL,
  `third_party_no` varchar(255) NULL,
  `paid_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_payment_no` (`payment_no`),
  INDEX `idx_order_id` (`order_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

-- Insert some test inventory data
INSERT INTO inventory (product_id, total, sold) VALUES
(1, 100, 0),
(2, 50, 0),
(3, 200, 0);

SET FOREIGN_KEY_CHECKS = 1;
