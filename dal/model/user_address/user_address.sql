-- 用户地址表 (PostgreSQL 版本)
CREATE TABLE user_addresses (
    address_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    detailed_address VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    province VARCHAR(100) DEFAULT NULL,
    is_default BOOLEAN DEFAULT false,
    recipient_name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(50) DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
