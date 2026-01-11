#!/bin/bash
set -e

echo "Testing TOKEN_SECRET persistence across server restarts..."
echo ""

BASE_URL="http://localhost:8080"

echo "Step 1: Creating a test event..."
EVENT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/events" \
  -H "X-Forwarded-User: admin" \
  -H "X-Forwarded-Email: admin@test.com" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Token Persistence Test",
    "description": "Testing invite tokens survive restart",
    "location": "Test Location",
    "start_time": "2026-02-01T18:00:00Z",
    "timezone": "America/Los_Angeles",
    "max_attendees": 50
  }')

EVENT_ID=$(echo "$EVENT_RESPONSE" | jq -r '.event.id // .id // empty')
if [ -z "$EVENT_ID" ]; then
  echo "❌ FAILED: Could not create event"
  echo "Response: $EVENT_RESPONSE"
  exit 1
fi
echo "Created event with ID: $EVENT_ID"
echo ""

echo "Step 2: Creating a manual invite..."
INVITE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/events/$EVENT_ID/invites/manual" \
  -H "X-Forwarded-User: admin" \
  -H "X-Forwarded-Email: admin@test.com" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Guest",
    "max_plus_ones": 2
  }')

TOKEN=$(echo "$INVITE_RESPONSE" | jq -r '.token // empty')
if [ -z "$TOKEN" ]; then
  echo "❌ FAILED: Could not create invite"
  echo "Response: $INVITE_RESPONSE"
  exit 1
fi
echo "Created invite with token: $TOKEN"
echo ""

echo "Step 3: Accessing invite BEFORE restart..."
BEFORE_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/rsvp/$TOKEN")
BEFORE_STATUS=$(echo "$BEFORE_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
echo "Response status: $BEFORE_STATUS"

if [ "$BEFORE_STATUS" != "200" ]; then
  echo "❌ FAILED: Invite not accessible before restart (status: $BEFORE_STATUS)"
  exit 1
fi
echo "✅ Invite accessible before restart"
echo ""

echo "Step 4: Restarting TinyRSVP container..."
docker compose -f docker-compose.test.yml restart tinyrsvp
echo "Waiting for service to be healthy..."
sleep 5

for i in {1..30}; do
  if curl -s "$BASE_URL/health" > /dev/null 2>&1; then
    echo "✅ Service is healthy"
    break
  fi
  if [ $i -eq 30 ]; then
    echo "❌ Service did not become healthy in time"
    exit 1
  fi
  sleep 1
done
echo ""

echo "Step 5: Accessing invite AFTER restart..."
AFTER_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/rsvp/$TOKEN")
AFTER_STATUS=$(echo "$AFTER_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
echo "Response status: $AFTER_STATUS"

if [ "$AFTER_STATUS" != "200" ]; then
  echo "❌ FAILED: Invite not accessible after restart (status: $AFTER_STATUS)"
  echo ""
  echo "Response body:"
  echo "$AFTER_RESPONSE" | grep -v "HTTP_STATUS"
  exit 1
fi

echo "✅ Invite accessible after restart"
echo ""
echo "=========================================="
echo "✅ ALL TESTS PASSED"
echo "=========================================="
echo "Token persistence verified across restart!"
