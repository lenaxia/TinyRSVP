#!/bin/bash

set -e

echo "Running Go vulnerability check..."

mkdir -p security-reports

govulncheck -json ./... > security-reports/govulncheck-report.json 2>&1 || true

echo "govulncheck scan complete! Report: security-reports/govulncheck-report.json"

if [ -f security-reports/govulncheck-report.json ]; then
  VULNS=$(grep -c '"osv"' security-reports/govulncheck-report.json 2>/dev/null || echo "0")
  echo "Found $VULNS vulnerabilities"
  
  if [ "$VULNS" != "0" ]; then
    echo "Vulnerabilities detected. Review security-reports/govulncheck-report.json"
    govulncheck ./...
    exit 1
  fi
fi

echo "No vulnerabilities found!"
