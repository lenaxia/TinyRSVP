#!/bin/bash
export SERVER_BASE_URL=http://localhost:8080
export DATABASE_PATH=./test.db
export SMTP_HOST=localhost
export SMTP_PORT=25
export EMAIL_FROM=test@example.com

./server &
SERVER_PID=$!
sleep 3

echo "Testing /health endpoint:"
curl -s http://localhost:8080/health | jq .

echo -e "\nTesting /ready endpoint:"
curl -s http://localhost:8080/ready | jq .

kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null
rm -f ./test.db
