#!/bin/bash

set -e

echo "=========================================="
echo "TinyRSVP Security Scan Suite"
echo "=========================================="
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"

mkdir -p security-reports

FAILED=0

echo "1. Running gosec static analysis..."
echo "------------------------------------------"
if bash "$SCRIPT_DIR/gosec-scan.sh"; then
  echo "✓ gosec passed"
else
  echo "✗ gosec failed"
  FAILED=$((FAILED + 1))
fi
echo ""

echo "2. Running Go vulnerability check..."
echo "------------------------------------------"
if bash "$SCRIPT_DIR/govulncheck-scan.sh"; then
  echo "✓ govulncheck passed"
else
  echo "✗ govulncheck failed"
  FAILED=$((FAILED + 1))
fi
echo ""

echo "=========================================="
echo "Security Scan Summary"
echo "=========================================="
echo "Total scans run: 2"
echo "Failed scans: $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
  echo "✓ All security scans passed!"
  echo "Reports available in: security-reports/"
  exit 0
else
  echo "✗ $FAILED security scan(s) failed"
  echo "Review reports in: security-reports/"
  exit 1
fi
