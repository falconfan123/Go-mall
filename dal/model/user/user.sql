-- 用户表 (PostgreSQL 版本)
CREATE TABLE users
(
    user_id       SERIAL PRIMARY KEY,
    username      VARCHAR(255) DEFAULT NULL,
    email         VARCHAR(255) UNIQUE,
    password_hash VARCHAR(512),
    avatar_url    VARCHAR(255) DEFAULT NULL,
    created_at    TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    user_deleted  BOOLEAN DEFAULT FALSE,
    logout_at     TIMESTAMP     DEFAULT NULL,
    login_at      TIMESTAMP     DEFAULT NULL,
    updated_at    TIMESTAMP     DEFAULT CURRENT_TIMESTAMP
);
