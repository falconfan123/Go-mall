#!/bin/bash

set -e

echo "=== Starting go-mall ==="
cd /Users/fan/go-mall

# Cleanup
echo "Cleaning up..."
pkill -f "go run" 2>/dev/null || true
pkill -f "services/" 2>/dev/null || true
pkill -f "apis/" 2>/dev/null || true
lsof -ti:10000,10001,10002,8001,8002 2>/dev/null | xargs kill -9 2>/dev/null || true
sleep 2

# Start backend RPC services
echo "Starting backend RPC services..."

cd /Users/fan/go-mall/services/auths
go run auths.go > /tmp/auths.log 2>&1 &
AUTHS_PID=$!
echo "  - auths.rpc (PID $AUTHS_PID)"
sleep 2

cd /Users/fan/go-mall/services/users
go run users.go > /tmp/users.log 2>&1 &
USERS_PID=$!
echo "  - users.rpc (PID $USERS_PID)"
sleep 2

cd /Users/fan/go-mall/services/product
go run product.go > /tmp/product.log 2>&1 &
PRODUCT_PID=$!
echo "  - products.rpc (PID $PRODUCT_PID)"
sleep 2

# Start API gateways
echo "Starting API gateways..."

cd /Users/fan/go-mall/apis/user
go run user.go > /tmp/user-api.log 2>&1 &
USER_API_PID=$!
echo "  - user-api (PID $USER_API_PID) on http://localhost:8001"
sleep 2

cd /Users/fan/go-mall/apis/product
go run product.go > /tmp/product-api.log 2>&1 &
PRODUCT_API_PID=$!
echo "  - product-api (PID $PRODUCT_API_PID) on http://localhost:8002"
sleep 2

echo ""
echo "=== Services started successfully! ==="
echo ""
echo "RPC Services:"
echo "  - auths.rpc:    10000"
echo "  - users.rpc:    10001"
echo "  - products.rpc: 10002"
echo ""
echo "API Endpoints:"
echo "  - User API:     http://localhost:8001"
echo "  - Product API:  http://localhost:8002"
echo ""
echo "Test URLs:"
echo "  - Product List: http://localhost:8002/douyin/product/list?page=1&size=10"
echo ""
echo "Consul UI: http://localhost:8500"
echo ""
echo "Logs:"
echo "  tail -f /tmp/auths.log"
echo "  tail -f /tmp/users.log"
echo "  tail -f /tmp/product.log"
echo "  tail -f /tmp/user-api.log"
echo "  tail -f /tmp/product-api.log"
echo ""
echo "To stop: pkill -f 'go run'"
echo ""

# Save PIDs
cat > /tmp/go-mall-pids.txt <<EOF
AUTHS_PID=$AUTHS_PID
USERS_PID=$USERS_PID
PRODUCT_PID=$PRODUCT_PID
USER_API_PID=$USER_API_PID
PRODUCT_API_PID=$PRODUCT_API_PID
EOF

# Wait a bit and test
echo "Waiting for services to be ready..."
sleep 5

echo ""
echo "=== Testing product API ==="
curl -s "http://localhost:8002/douyin/product/list?page=1&size=10" 2>/dev/null || echo "API not ready yet, check logs"

echo ""
echo ""
echo "=== Consul Services ==="
curl -s http://localhost:8500/v1/catalog/services 2>/dev/null || echo "Consul not reachable"

echo ""
echo "=== Done! ==="
