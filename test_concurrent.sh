#!/bin/bash
export SERVER_BASE_URL=http://localhost:8080
export DATABASE_PATH=./test_concurrent.db
export SMTP_HOST=localhost
export SMTP_PORT=25
export EMAIL_FROM=test@example.com

./server > /dev/null 2>&1 &
SERVER_PID=$!
sleep 3

echo "Sending 20 concurrent requests to /health..."
for i in {1..20}; do
  curl -s http://localhost:8080/health > /dev/null &
done
wait
echo "All /health requests completed"

echo "Sending 20 concurrent requests to /ready..."
for i in {1..20}; do
  curl -s http://localhost:8080/ready > /dev/null &
done
wait
echo "All /ready requests completed"

kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null
rm -f ./test_concurrent.db
