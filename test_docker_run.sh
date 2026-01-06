#!/bin/bash

set -e

echo "=== Docker Runtime Tests ==="
echo ""

echo "1. Starting container with docker compose..."
docker compose up -d

echo ""
echo "2. Waiting for container to be healthy (max 60s)..."
TIMEOUT=60
ELAPSED=0
while [ $ELAPSED -lt $TIMEOUT ]; do
    if docker compose ps | grep -q "healthy"; then
        echo "   ✓ Container is healthy"
        break
    fi
    sleep 2
    ELAPSED=$((ELAPSED + 2))
    echo -n "."
done

if [ $ELAPSED -ge $TIMEOUT ]; then
    echo ""
    echo "   ❌ FAIL: Container did not become healthy within ${TIMEOUT}s"
    docker compose logs tinyrsvp
    docker compose down
    exit 1
fi

echo ""
echo "3. Checking health endpoint..."
if curl -f -s http://localhost:8080/health > /dev/null; then
    echo "   ✓ PASS: Health endpoint responding"
else
    echo "   ❌ FAIL: Health endpoint not responding"
    docker compose logs tinyrsvp
    docker compose down
    exit 1
fi

echo ""
echo "4. Checking readiness endpoint..."
if curl -f -s http://localhost:8080/ready > /dev/null; then
    echo "   ✓ PASS: Readiness endpoint responding"
else
    echo "   ❌ FAIL: Readiness endpoint not responding"
    docker compose logs tinyrsvp
    docker compose down
    exit 1
fi

echo ""
echo "5. Verifying database file created..."
if docker compose exec -T tinyrsvp ls /data/tinyrsvp.db > /dev/null 2>&1; then
    echo "   ✓ PASS: Database file exists"
else
    echo "   ❌ FAIL: Database file not found"
    docker compose logs tinyrsvp
    docker compose down
    exit 1
fi

echo ""
echo "6. Checking container logs for errors..."
if docker compose logs tinyrsvp | grep -i "error\|fatal\|panic" | grep -v "test"; then
    echo "   ⚠ WARNING: Found error messages in logs"
else
    echo "   ✓ PASS: No errors in logs"
fi

echo ""
echo "7. Stopping container..."
docker compose down

echo ""
echo "8. Testing volume persistence..."
docker compose up -d
sleep 5

if docker compose exec -T tinyrsvp ls /data/tinyrsvp.db > /dev/null 2>&1; then
    echo "   ✓ PASS: Database persisted after restart"
else
    echo "   ❌ FAIL: Database not persisted"
    docker compose down
    exit 1
fi

echo ""
echo "9. Final cleanup..."
docker compose down

echo ""
echo "=== All runtime tests passed! ==="
