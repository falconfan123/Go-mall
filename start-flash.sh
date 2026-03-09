#!/bin/bash

echo "=== Starting Go-Mall Flash Sale ==="
cd /Users/fan/go-mall

# Cleanup
echo "Cleaning up..."
pkill -f "go run" 2>/dev/null || true
lsof -ti:10000,10001,8001,8008 2>/dev/null | xargs kill -9 2>/dev/null || true
sleep 2

# Start core services (auths, users)
echo "Starting core services..."

cd /Users/fan/go-mall/services/auths
go run auths.go > /tmp/auths.log 2>&1 &
AUTHS_PID=$!
echo "  - auths.rpc (PID $AUTHS_PID)"
sleep 3

cd /Users/fan/go-mall/services/users
go run users.go > /tmp/users.log 2>&1 &
USERS_PID=$!
echo "  - users.rpc (PID $USERS_PID)"
sleep 3

# Start product service
cd /Users/fan/go-mall/services/product
go run product.go > /tmp/product.log 2>&1 &
PRODUCT_PID=$!
echo "  - products.rpc (PID $PRODUCT_PID)"
sleep 3

# Start user API
cd /Users/fan/go-mall/apis/user
go run user.go > /tmp/user-api.log 2>&1 &
USER_API_PID=$!
echo "  - user-api (PID $USER_API_PID) on http://localhost:8001"
sleep 3

# Start product API
cd /Users/fan/go-mall/apis/product
go run product.go > /tmp/product-api.log 2>&1 &
PRODUCT_API_PID=$!
echo "  - product-api (PID $PRODUCT_API_PID) on http://localhost:8002"
sleep 3

# Start flash sale API
cd /Users/fan/go-mall/apis/flash_sale
go run flash.go > /tmp/flash-api.log 2>&1 &
FLASH_API_PID=$!
echo "  - flash-api (PID $FLASH_API_PID) on http://localhost:8008"
sleep 3

# Start frontend server (if not already running)
echo "Starting frontend server..."
cd /Users/fan/go-mall/frontend
python3 -m http.server 3000 > /tmp/frontend.log 2>&1 &
FRONTEND_PID=$!
echo "  - frontend (PID $FRONTEND_PID) on http://localhost:3000"

echo ""
echo "=== Services started! ==="
echo ""
echo "RPC Services:"
echo "  - auths.rpc:    10000"
echo "  - users.rpc:    10001"
echo "  - products.rpc: 10002"
echo ""
echo "API Endpoints:"
echo "  - User API:     http://localhost:8001"
echo "  - Product API:  http://localhost:8002"
echo "  - Flash API:    http://localhost:8008"
echo ""
echo "Frontend:"
echo "  - Main Page:    http://localhost:3000"
echo "  - Flash Sale:   http://localhost:3000/#"
echo ""
echo "Test URLs:"
echo "  - Flash Products: http://localhost:8008/douyin/flash/products"
echo ""
echo "Consul UI: http://localhost:8500"
echo ""
echo "Logs:"
echo "  tail -f /tmp/auths.log"
echo "  tail -f /tmp/users.log"
echo "  tail -f /tmp/product.log"
echo "  tail -f /tmp/user-api.log"
echo "  tail -f /tmp/product-api.log"
echo "  tail -f /tmp/flash-api.log"
echo "  tail -f /tmp/frontend.log"
echo ""
echo "To stop: pkill -f 'go run' && pkill -f 'python3 -m http.server'"
echo ""

# Save PIDs
cat > /tmp/go-mall-pids.txt <<EOF
AUTHS_PID=$AUTHS_PID
USERS_PID=$USERS_PID
PRODUCT_PID=$PRODUCT_PID
USER_API_PID=$USER_API_PID
PRODUCT_API_PID=$PRODUCT_API_PID
FLASH_API_PID=$FLASH_API_PID
FRONTEND_PID=$FRONTEND_PID
EOF

echo "Waiting for services to be ready..."
sleep 5

echo ""
echo "=== Consul Services ==="
curl -s http://localhost:8500/v1/catalog/services 2>/dev/null || echo "Consul not reachable"

echo ""
echo "=== Log snippet: auths ==="
tail -15 /tmp/auths.log 2>/dev/null || echo "No log"

echo ""
echo "=== Log snippet: users ==="
tail -15 /tmp/users.log 2>/dev/null || echo "No log"

echo ""
echo "=== Log snippet: product ==="
tail -15 /tmp/product.log 2>/dev/null || echo "No log"

echo ""
echo "=== Log snippet: user-api ==="
tail -15 /tmp/user-api.log 2>/dev/null || echo "No log"

echo ""
echo "=== Log snippet: product-api ==="
tail -15 /tmp/product-api.log 2>/dev/null || echo "No log"

echo ""
echo "=== Log snippet: flash-api ==="
tail -15 /tmp/flash-api.log 2>/dev/null || echo "No log"

echo ""
echo "=== Done! ==="
