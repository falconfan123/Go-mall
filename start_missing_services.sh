#!/bin/bash

# Function to start a service if not running
start_service() {
    service_name=$1
    service_path=$2
    
    if ps aux | grep "go run $service_name.go" | grep -v grep > /dev/null; then
        echo "$service_name is already running."
    else
        echo "Starting $service_name..."
        cd $service_path
        nohup go run $service_name.go > $service_name.log 2>&1 &
        cd - > /dev/null
    fi
}

start_service "product" "services/product"
start_service "carts" "services/carts"
start_service "payment" "services/payment"
start_service "inventory" "services/inventory"
start_service "coupons" "services/coupons"

echo "All missing services started."
