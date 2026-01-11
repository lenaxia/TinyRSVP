#!/bin/bash

set -e

echo "Running gosec static security analysis..."

mkdir -p security-reports

gosec \
  -fmt=json \
  -out=security-reports/gosec-report.json \
  -exclude-dir=tests \
  -exclude-dir=scripts \
  ./...

echo "gosec scan complete! Report: security-reports/gosec-report.json"

if [ -f security-reports/gosec-report.json ]; then
  ISSUES=$(jq '.Issues | length' security-reports/gosec-report.json 2>/dev/null || echo "0")
  echo "Found $ISSUES security issues"
  
  if [ "$ISSUES" != "0" ]; then
    echo "Security issues detected. Review security-reports/gosec-report.json"
    exit 1
  fi
fi

echo "No security issues found!"
