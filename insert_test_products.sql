USE mall;

-- Insert test products (skip if already exists)
INSERT IGNORE INTO products (id, NAME, description, picture, price, stock) VALUES
(1, 'iPhone 15', '最新款苹果手机，搭载 A17 芯片', 'https://example.com/iphone15.jpg', 7999, 100),
(2, 'MacBook Pro', '高性能笔记本电脑，M3 Pro 芯片', 'https://example.com/macbook.jpg', 14999, 50),
(3, 'Nike Air Max', '舒适的运动鞋子，气垫设计', 'https://example.com/shoes.jpg', 899, 200),
(4, 'Sony WH-1000XM5', '顶级降噪耳机', 'https://example.com/headphones.jpg', 2699, 75),
(5, 'iPad Pro', '12.9英寸平板电脑，M2芯片', 'https://example.com/ipad.jpg', 8999, 40);

-- Link products to categories
INSERT IGNORE INTO product_categories (product_id, category_id) VALUES
(1, 1),
(2, 1),
(3, 2),
(4, 1),
(4, 3),
(5, 1);
