# User Story: Security Scanning Setup

**Epic:** 09 - Security Review & Penetration Testing  
**Priority:** Critical  
**Status:** Not Started  
**Estimated Effort:** 1 day

---

## Story

As a **security engineer**, I want to **set up automated security scanning tools** so that **vulnerabilities can be detected early and continuously throughout development**.

---

## Acceptance Criteria

- [ ] OWASP ZAP configured and integrated
- [ ] Nuclei vulnerability scanner installed and configured
- [ ] gosec static analysis tool integrated
- [ ] Trivy container and dependency scanner configured
- [ ] govulncheck Go vulnerability checker integrated
- [ ] Security scanning can run in CI/CD pipeline
- [ ] Scan results exported in machine-readable format
- [ ] Baseline scan completed with zero critical issues
- [ ] Documentation created for running security scans

---

## Tasks

### OWASP ZAP Setup
- [ ] Pull OWASP ZAP Docker image
- [ ] Create ZAP configuration file
- [ ] Configure authentication for authenticated scans
- [ ] Set up baseline scan script
- [ ] Set up full scan script
- [ ] Configure report generation (HTML, JSON, XML)
- [ ] Test scan against running application

### Nuclei Setup
- [ ] Install Nuclei binary or Docker image
- [ ] Update Nuclei templates to latest
- [ ] Create custom templates for TinyRSVP-specific tests
- [ ] Configure scan targets and exclusions
- [ ] Set up severity filtering
- [ ] Test scan execution

### gosec Setup
- [ ] Install gosec tool
- [ ] Create gosec configuration file (.gosec.json)
- [ ] Configure rules and exclusions
- [ ] Set up JSON output for parsing
- [ ] Integrate with go test workflow
- [ ] Run initial scan and establish baseline

### Trivy Setup
- [ ] Install Trivy scanner
- [ ] Configure container image scanning
- [ ] Configure filesystem scanning
- [ ] Configure dependency scanning (go.mod)
- [ ] Set up severity thresholds
- [ ] Configure ignore policies for false positives
- [ ] Test all scan types

### govulncheck Setup
- [ ] Install govulncheck tool
- [ ] Run initial vulnerability check
- [ ] Document any found vulnerabilities
- [ ] Create remediation plan for vulnerabilities
- [ ] Integrate into CI/CD pipeline

### CI/CD Integration
- [ ] Create security-scan.sh script
- [ ] Add security scan stage to CI/CD pipeline
- [ ] Configure failure thresholds
- [ ] Set up scan result artifacts
- [ ] Configure notifications for security findings

---

## Technical Details

### OWASP ZAP Configuration

**Baseline Scan:**
```bash
#!/bin/bash
# scripts/security/zap-baseline.sh

docker run -t owasp/zap2docker-stable \
  zap-baseline.py \
  -t http://localhost:8080 \
  -r zap-baseline-report.html \
  -J zap-baseline-report.json \
  -w zap-baseline-report.md
```

**Full Scan:**
```bash
#!/bin/bash
# scripts/security/zap-full.sh

docker run -t owasp/zap2docker-stable \
  zap-full-scan.py \
  -t http://localhost:8080 \
  -r zap-full-report.html \
  -J zap-full-report.json \
  -w zap-full-report.md
```

**Authenticated Scan:**
```bash
#!/bin/bash
# scripts/security/zap-auth.sh

# Create ZAP context with authentication
docker run -v $(pwd)/zap-config:/zap/wrk:rw \
  -t owasp/zap2docker-stable \
  zap-full-scan.py \
  -t http://localhost:8080 \
  -n /zap/wrk/context.context \
  -r zap-auth-report.html
```

### Nuclei Configuration

**Installation:**
```bash
go install -v github.com/projectdiscovery/nuclei/v2/cmd/nuclei@latest
nuclei -update-templates
```

**Scan Script:**
```bash
#!/bin/bash
# scripts/security/nuclei-scan.sh

nuclei \
  -u http://localhost:8080 \
  -t nuclei-templates/ \
  -severity critical,high,medium \
  -o nuclei-results.txt \
  -json-export nuclei-results.json
```

**Custom Template Example:**
```yaml
# scripts/security/nuclei-custom/tinyrsvp-token-exposure.yaml
id: tinyrsvp-token-exposure

info:
  name: TinyRSVP Token Exposure
  author: security-team
  severity: high
  description: Checks for invite token exposure in error messages

requests:
  - method: GET
    path:
      - "{{BaseURL}}/rsvp/invalid-token-test"
    
    matchers:
      - type: regex
        regex:
          - "[a-zA-Z0-9_-]{43}"
        part: body
```

### gosec Configuration

**Configuration File:**
```json
{
  "global": {
    "nosec": false,
    "audit": true
  },
  "severity": "medium",
  "confidence": "medium",
  "exclude": [
    "G104"
  ]
}
```

**Scan Script:**
```bash
#!/bin/bash
# scripts/security/gosec-scan.sh

gosec \
  -fmt=json \
  -out=gosec-report.json \
  -exclude-dir=tests \
  ./...
```

### Trivy Configuration

**Container Scan:**
```bash
#!/bin/bash
# scripts/security/trivy-container.sh

trivy image \
  --severity HIGH,CRITICAL \
  --format json \
  --output trivy-container-report.json \
  tinyrsvp:latest
```

**Filesystem Scan:**
```bash
#!/bin/bash
# scripts/security/trivy-fs.sh

trivy fs \
  --severity HIGH,CRITICAL \
  --format json \
  --output trivy-fs-report.json \
  .
```

**Dependency Scan:**
```bash
#!/bin/bash
# scripts/security/trivy-deps.sh

trivy fs \
  --scanners vuln \
  --severity HIGH,CRITICAL \
  --format json \
  --output trivy-deps-report.json \
  go.mod
```

### govulncheck Integration

**Scan Script:**
```bash
#!/bin/bash
# scripts/security/govulncheck-scan.sh

govulncheck -json ./... > govulncheck-report.json
```

### Master Security Scan Script

```bash
#!/bin/bash
# scripts/security/run-all-scans.sh

set -e

echo "Starting comprehensive security scan..."

# Create reports directory
mkdir -p security-reports

# Run gosec
echo "Running gosec static analysis..."
gosec -fmt=json -out=security-reports/gosec-report.json ./...

# Run govulncheck
echo "Running Go vulnerability check..."
govulncheck -json ./... > security-reports/govulncheck-report.json

# Build Docker image for scanning
echo "Building Docker image..."
docker build -t tinyrsvp:security-test .

# Run Trivy scans
echo "Running Trivy container scan..."
trivy image --severity HIGH,CRITICAL \
  --format json \
  --output security-reports/trivy-container.json \
  tinyrsvp:security-test

echo "Running Trivy filesystem scan..."
trivy fs --severity HIGH,CRITICAL \
  --format json \
  --output security-reports/trivy-fs.json \
  .

# Start application for dynamic testing
echo "Starting application..."
docker run -d --name tinyrsvp-test \
  -p 8080:8080 \
  -e DATABASE_PATH=/tmp/test.db \
  tinyrsvp:security-test

# Wait for application to be ready
sleep 5

# Run OWASP ZAP baseline
echo "Running OWASP ZAP baseline scan..."
docker run -t owasp/zap2docker-stable \
  zap-baseline.py \
  -t http://host.docker.internal:8080 \
  -r security-reports/zap-baseline.html \
  -J security-reports/zap-baseline.json

# Run Nuclei
echo "Running Nuclei vulnerability scan..."
nuclei -u http://localhost:8080 \
  -t nuclei-templates/ \
  -severity critical,high,medium \
  -json-export security-reports/nuclei-results.json

# Cleanup
echo "Cleaning up..."
docker stop tinyrsvp-test
docker rm tinyrsvp-test

echo "Security scan complete! Reports in security-reports/"
```

---

## Testing Strategy

### Unit Tests
```go
// tests/security/scanner_test.go
package security_test

import (
	"os/exec"
	"testing"
)

func TestGosecInstalled(t *testing.T) {
	cmd := exec.Command("gosec", "-version")
	if err := cmd.Run(); err != nil {
		t.Fatal("gosec not installed")
	}
}

func TestTrivyInstalled(t *testing.T) {
	cmd := exec.Command("trivy", "--version")
	if err := cmd.Run(); err != nil {
		t.Fatal("trivy not installed")
	}
}

func TestNucleiInstalled(t *testing.T) {
	cmd := exec.Command("nuclei", "-version")
	if err := cmd.Run(); err != nil {
		t.Fatal("nuclei not installed")
	}
}
```

### Integration Tests
- [ ] Run full security scan suite
- [ ] Verify all tools execute successfully
- [ ] Validate report generation
- [ ] Check for critical findings
- [ ] Verify CI/CD integration

---

## Security Baseline

### Expected Results (Initial Scan)
- **gosec:** 0 high-severity issues
- **Trivy:** 0 critical vulnerabilities in dependencies
- **OWASP ZAP:** 0 high-risk alerts
- **Nuclei:** 0 critical findings
- **govulncheck:** 0 known vulnerabilities

### Acceptable Findings
- Low-severity informational findings
- False positives (documented and justified)
- Known issues with remediation plan

### Unacceptable Findings
- Critical or high-severity vulnerabilities
- SQL injection vulnerabilities
- XSS vulnerabilities
- Authentication bypass
- Authorization bypass
- Sensitive data exposure

---

## Documentation

### Security Scanning Guide
Create `docs/SECURITY_SCANNING.md`:
- Tool installation instructions
- How to run each scanner
- How to interpret results
- How to handle false positives
- Remediation workflow

### CI/CD Integration Guide
- Pipeline configuration
- Automated scan triggers
- Failure thresholds
- Notification setup

---

## Dependencies

**Depends on:** None (can start immediately)  
**Blocks:** All other security stories

---

## Definition of Done

- [ ] All security tools installed and configured
- [ ] Baseline scans completed
- [ ] No critical or high-severity findings
- [ ] All scan scripts tested and working
- [ ] CI/CD integration complete
- [ ] Documentation complete
- [ ] Team trained on running scans
- [ ] Security baseline established

---

## References

- OWASP ZAP: https://www.zaproxy.org/
- Nuclei: https://github.com/projectdiscovery/nuclei
- gosec: https://github.com/securego/gosec
- Trivy: https://github.com/aquasecurity/trivy
- govulncheck: https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck
