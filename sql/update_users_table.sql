USE mall;

-- Drop and recreate the users table with all required columns
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    user_id INT NOT NULL AUTO_INCREMENT COMMENT '主键，自增，用户 ID',
    username VARCHAR(255) NULL DEFAULT NULL COMMENT '用户名，可空',
    email VARCHAR(255) NULL DEFAULT NULL COMMENT '邮箱，唯一',
    password_hash VARCHAR(512) NULL DEFAULT NULL COMMENT '密码哈希值',
    avatar_url VARCHAR(255) NULL DEFAULT NULL COMMENT '头像图片 URL',
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    user_deleted TINYINT(1) NULL DEFAULT 0 COMMENT '用户是否已删除',
    logout_at TIMESTAMP NULL DEFAULT NULL COMMENT '最近一次登出时间',
    login_at TIMESTAMP NULL DEFAULT NULL COMMENT '最近一次登录时间',
    updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (user_id) USING BTREE,
    UNIQUE INDEX email (email ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- Verify the structure
DESCRIBE users;
