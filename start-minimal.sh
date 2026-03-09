#!/bin/bash

echo "=== Starting go-mall (minimal) ==="
cd /Users/fan/go-mall

# Cleanup
echo "Cleaning up..."
pkill -f "go run" 2>/dev/null || true
lsof -ti:10000,10001,8001 2>/dev/null | xargs kill -9 2>/dev/null || true
sleep 2

# Start auths service
echo "Starting auths service..."
cd /Users/fan/go-mall/services/auths
go run auths.go > /tmp/auths.log 2>&1 &
AUTHS_PID=$!
echo "  - auths.rpc (PID $AUTHS_PID)"
sleep 3

# Start users service
echo "Starting users service..."
cd /Users/fan/go-mall/services/users
go run users.go > /tmp/users.log 2>&1 &
USERS_PID=$!
echo "  - users.rpc (PID $USERS_PID)"
sleep 3

# Start user API
echo "Starting user API..."
cd /Users/fan/go-mall/apis/user
go run user.go > /tmp/user-api.log 2>&1 &
USER_API_PID=$!
echo "  - user-api (PID $USER_API_PID) on http://localhost:8001"
sleep 3

echo ""
echo "=== Services started! ==="
echo ""
echo "RPC Services:"
echo "  - auths.rpc: 10000"
echo "  - users.rpc: 10001"
echo ""
echo "API Endpoints:"
echo "  - User API: http://localhost:8001"
echo ""
echo "Test URLs:"
echo "  - Register: POST http://localhost:8001/douyin/user/register"
echo "  - Login:    POST http://localhost:8001/douyin/user/login"
echo ""
echo "Consul UI: http://localhost:8500"
echo ""
echo "Logs:"
echo "  tail -f /tmp/auths.log"
echo "  tail -f /tmp/users.log"
echo "  tail -f /tmp/user-api.log"
echo ""
echo "To stop: pkill -f 'go run'"
echo ""

# Save PIDs
cat > /tmp/go-mall-pids.txt <<EOF
AUTHS_PID=$AUTHS_PID
USERS_PID=$USERS_PID
USER_API_PID=$USER_API_PID
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
echo "=== Log snippet: user-api ==="
tail -15 /tmp/user-api.log 2>/dev/null || echo "No log"

echo ""
echo "=== Done! ==="
