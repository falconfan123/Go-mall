-- 库存表 (PostgreSQL 版本)
CREATE TABLE inventory
(
    product_id INTEGER PRIMARY KEY,
    total      INTEGER NOT NULL,
    sold       INTEGER NOT NULL
);
