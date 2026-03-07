#!/bin/bash

# Test script for GoFoundation SDK

echo "Building basic example..."
cd examples/basic
go build -o basic

echo "Starting server in background..."
./basic &
SERVER_PID=$!

# Wait for server to start
sleep 2

echo "Testing /api/users endpoint..."
RESPONSE=$(curl -s http://localhost:8080/api/users)
echo "Response: $RESPONSE"

# Check if response contains expected fields
if echo "$RESPONSE" | grep -q "trace_id" && echo "$RESPONSE" | grep -q "data"; then
    echo "✓ Response format is correct"
else
    echo "✗ Response format is incorrect"
fi

echo "Testing /api/health endpoint..."
HEALTH=$(curl -s http://localhost:8080/api/health)
echo "Response: $HEALTH"

# Stop server
kill $SERVER_PID

echo "Test completed!"
