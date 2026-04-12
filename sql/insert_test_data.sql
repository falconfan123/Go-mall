USE mall;

-- Insert test categories
INSERT INTO categories (NAME, description) VALUES
('电子产品', '各种电子设备和配件'),
('服装', '时尚服装和配饰'),
('食品', '各种食品和饮料');

-- Insert test products
INSERT INTO products (NAME, description, picture, price, stock) VALUES
('iPhone 15', '最新款苹果手机', 'https://example.com/iphone15.jpg', 7999.00, 100),
('MacBook Pro', '高性能笔记本电脑', 'https://example.com/macbook.jpg', 14999.00, 50),
('Nike运动鞋', '舒适的运动鞋子', 'https://example.com/shoes.jpg', 899.00, 200);

-- Insert test user (password is "password123" hashed)
-- The hash is for "password123" using bcrypt
INSERT INTO users (username, email, password_hash, avatar_url) VALUES
('testuser', 'test@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhW', 'https://example.com/avatar.jpg');
