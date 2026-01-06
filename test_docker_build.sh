#!/bin/bash

set -e

echo "=== Docker Build Tests ==="
echo ""

echo "1. Building Docker image..."
docker build -t tinyrsvp:test .

echo ""
echo "2. Checking image size..."
SIZE=$(docker images tinyrsvp:test --format "{{.Size}}")
echo "   Image size: $SIZE"

SIZE_MB=$(docker images tinyrsvp:test --format "{{.Size}}" | sed 's/MB//' | awk '{print int($1)}')
if [ "$SIZE_MB" -gt 50 ]; then
    echo "   ❌ FAIL: Image size exceeds 50MB"
    exit 1
fi
echo "   ✓ PASS: Image size is under 50MB"

echo ""
echo "3. Verifying binary exists..."
docker run --rm tinyrsvp:test ls -lh /app/tinyrsvp
echo "   ✓ PASS: Binary exists"

echo ""
echo "4. Verifying migrations directory exists..."
docker run --rm tinyrsvp:test ls -lh /app/migrations
echo "   ✓ PASS: Migrations directory exists"

echo ""
echo "5. Verifying templates directory exists..."
docker run --rm tinyrsvp:test ls -lh /app/templates
echo "   ✓ PASS: Templates directory exists"

echo ""
echo "6. Verifying static directory exists..."
docker run --rm tinyrsvp:test ls -lh /app/static
echo "   ✓ PASS: Static directory exists"

echo ""
echo "7. Verifying non-root user..."
USER_INFO=$(docker run --rm tinyrsvp:test id)
echo "   User info: $USER_INFO"
if echo "$USER_INFO" | grep -q "uid=1000(tinyrsvp)"; then
    echo "   ✓ PASS: Running as non-root user tinyrsvp"
else
    echo "   ❌ FAIL: Not running as expected user"
    exit 1
fi

echo ""
echo "8. Verifying data directory exists..."
docker run --rm tinyrsvp:test ls -lhd /data
echo "   ✓ PASS: Data directory exists"

echo ""
echo "=== All build tests passed! ==="
