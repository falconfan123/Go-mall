#!/bin/bash

echo "=== Starting go-mall services ==="
cd /Users/fan/go-mall

# Kill any existing processes on our ports
echo "Cleaning up existing processes..."
pkill -f "go run" 2>/dev/null || true
pkill -f "services/" 2>/dev/null || true

# Start auths service first
echo "Starting auths service..."
cd /Users/fan/go-mall/services/auths
go run auths.go > /tmp/auths.log 2>&1 &
AUTHS_PID=$!
echo "Auths PID: $AUTHS_PID"
sleep 3

# Start users service
echo "Starting users service..."
cd /Users/fan/go-mall/services/users
go run users.go > /tmp/users.log 2>&1 &
USERS_PID=$!
echo "Users PID: $USERS_PID"
sleep 3

# Start product service
echo "Starting product service..."
cd /Users/fan/go-mall/services/product
go run product.go > /tmp/product.log 2>&1 &
PRODUCT_PID=$!
echo "Product PID: $PRODUCT_PID"
sleep 3

# Start user API
echo "Starting user API..."
cd /Users/fan/go-mall/apis/user
go run user.go > /tmp/user-api.log 2>&1 &
USER_API_PID=$!
echo "User API PID: $USER_API_PID"
sleep 3

# Start product API
echo "Starting product API..."
cd /Users/fan/go-mall/apis/product
go run product.go > /tmp/product-api.log 2>&1 &
PRODUCT_API_PID=$!
echo "Product API PID: $PRODUCT_API_PID"
sleep 3

echo ""
echo "=== Services started ==="
echo "Auths PID: $AUTHS_PID"
echo "Users PID: $USERS_PID"
echo "Product PID: $PRODUCT_PID"
echo "User API PID: $USER_API_PID"
echo "Product API PID: $PRODUCT_API_PID"
echo ""
echo "Logs available at:"
echo "  /tmp/auths.log"
echo "  /tmp/users.log"
echo "  /tmp/product.log"
echo "  /tmp/user-api.log"
echo "  /tmp/product-api.log"
echo ""
echo "Checking service status..."
sleep 2

echo ""
echo "=== Log tail ==="
echo "--- Auths ---"
tail -20 /tmp/auths.log || true
echo ""
echo "--- Users ---"
tail -20 /tmp/users.log || true
echo ""
echo "--- Product ---"
tail -20 /tmp/product.log || true
echo ""
echo "--- User API ---"
tail -20 /tmp/user-api.log || true
echo ""
echo "--- Product API ---"
tail -20 /tmp/product-api.log || true

echo ""
echo "=== Consul services ==="
curl -s http://localhost:8500/v1/catalog/services | python3 -m json.tool 2>/dev/null || curl -s http://localhost:8500/v1/catalog/services

echo ""
echo "=== To stop all services, run: ==="
echo "pkill -f 'go run'"
echo ""
echo "PIDs saved to /tmp/go-mall-pids.txt"
echo "$AUTHS_PID" > /tmp/go-mall-pids.txt
echo "$USERS_PID" >> /tmp/go-mall-pids.txt
echo "$PRODUCT_PID" >> /tmp/go-mall-pids.txt
echo "$USER_API_PID" >> /tmp/go-mall-pids.txt
echo "$PRODUCT_API_PID" >> /tmp/go-mall-pids.txt
