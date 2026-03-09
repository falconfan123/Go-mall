#!/bin/bash
mkdir -p logs

echo "Starting Auths..."
(cd services/auths && nohup go run auths.go > ../../logs/auths.log 2>&1 &)

echo "Starting Checkout..."
(cd services/checkout && nohup go run checkout.go > ../../logs/checkout.log 2>&1 &)

echo "Starting Order..."
(cd services/order && nohup go run order.go > ../../logs/order.log 2>&1 &)

echo "Starting Users..."
(cd services/users && nohup go run users.go > ../../logs/users.log 2>&1 &)

echo "Starting Product..."
(cd services/product && nohup go run product.go > ../../logs/product.log 2>&1 &)

echo "Starting Payment..."
(cd services/payment && nohup go run payment.go > ../../logs/payment.log 2>&1 &)

echo "Starting Inventory..."
(cd services/inventory && nohup go run inventory.go > ../../logs/inventory.log 2>&1 &)

echo "Starting Coupons..."
(cd services/coupons && nohup go run coupons.go > ../../logs/coupons.log 2>&1 &)

echo "Starting Carts..."
(cd services/carts && nohup go run carts.go > ../../logs/carts.log 2>&1 &)

echo "Waiting for RPC services to initialize..."
sleep 15

echo "Starting User API..."
(cd apis/user && nohup go run user.go > ../../logs/user-api.log 2>&1 &)

echo "Starting Product API..."
(cd apis/product && nohup go run product.go > ../../logs/product-api.log 2>&1 &)

echo "Starting Order API..."
(cd apis/order && nohup go run order.go > ../../logs/order-api.log 2>&1 &)

echo "Starting Checkout API..."
(cd apis/checkout && nohup go run checkout.go > ../../logs/checkout-api.log 2>&1 &)

echo "Starting Payment API..."
(cd apis/payment && nohup go run payment.go > ../../logs/payment-api.log 2>&1 &)

echo "Starting Carts API..."
(cd apis/carts && nohup go run carts.go > ../../logs/carts-api.log 2>&1 &)

echo "Starting Coupon API..."
(cd apis/coupon && nohup go run coupon.go > ../../logs/coupon-api.log 2>&1 &)

echo "Starting Flash API..."
(cd apis/flash_sale && nohup go run flash.go > ../../logs/flash.log 2>&1 &)

echo "All services started!"
