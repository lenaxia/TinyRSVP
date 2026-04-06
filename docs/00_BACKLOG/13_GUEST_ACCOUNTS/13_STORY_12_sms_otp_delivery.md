# User Story: SMS OTP Delivery Implementation

**Epic:** [Epic 13 - Guest Accounts & Encryption at Rest](README.md)  
**Priority:** Low  
**Status:** Not Started  
**Estimated Effort:** 4 hours  

---

## User Story

As a **guest who registered with a phone number**, I want **to receive my OTP via SMS** so that **I can log in without needing an email address**.

---

## Acceptance Criteria

- [ ] `SMSOTPDelivery` implements the `OTPDelivery` interface from Story 08
- [ ] SMS is sent via a configurable HTTP provider (Twilio-compatible REST API)
- [ ] If no SMS provider is configured (`SMS_PROVIDER_URL` unset), `SMSOTPDelivery.Send` returns `ErrSMSNotConfigured` immediately — it does not panic or silently drop the message
- [ ] `EmailOTPDelivery.Send` (from Story 08) calls `SMSOTPDelivery` for phone identifiers when SMS is configured, falls back to `ErrSMSNotConfigured` when not
- [ ] The SMS message body contains only the 6-digit code and the app name — no invite links, no URLs
- [ ] Provider credentials (`SMS_PROVIDER_URL`, `SMS_FROM_NUMBER`, `SMS_AUTH_TOKEN`) are read from environment variables at construction time
- [ ] No SMS provider library is added as a Go dependency — all calls use `net/http`
- [ ] All tests use a mock HTTP server — no real SMS calls in tests
- [ ] All tests pass with timeout

---

## Technical Details

### File

```
internal/guestauth/delivery_sms.go
internal/guestauth/delivery_sms_test.go
```

### Config

```go
type SMSConfig struct {
    ProviderURL string // e.g. https://api.twilio.com/2010-04-01/Accounts/{SID}/Messages.json
    FromNumber  string // E.164 format e.g. +15551234567
    AuthToken   string // HTTP Basic Auth password (SID is in ProviderURL)
}

func SMSConfigFromEnv() (*SMSConfig, error)
// Returns nil, ErrSMSNotConfigured if SMS_PROVIDER_URL is unset
```

### SMSOTPDelivery

```go
type SMSOTPDelivery struct {
    cfg    *SMSConfig
    client *http.Client
}

func NewSMSOTPDelivery(cfg *SMSConfig) *SMSOTPDelivery

func (d *SMSOTPDelivery) Send(ctx context.Context, phone, code string) error
// POST to cfg.ProviderURL with form body: To=phone&From=cfg.FromNumber&Body=<message>
// Uses HTTP Basic Auth with SID (extracted from URL) and cfg.AuthToken
// Returns wrapped error on non-2xx response
```

### Message Format

```
Your TinyRSVP login code is: 123456

This code expires in 15 minutes.
```

No links, no invite tokens, no URLs.

### OTPDelivery Routing (update to EmailOTPDelivery from Story 08)

The `EmailOTPDelivery` introduced in Story 08 is updated to also hold a reference to `OTPDelivery` for SMS, routing by identifier type:

```go
type routingOTPDelivery struct {
    emailDelivery OTPDelivery
    smsDelivery   OTPDelivery // may be nil if SMS not configured
}

func NewOTPDelivery(emailSvc email.Service, smsCfg *SMSConfig) OTPDelivery

func (d *routingOTPDelivery) Send(ctx context.Context, identifier, code string) error {
    if isEmail(identifier) {
        return d.emailDelivery.Send(ctx, identifier, code)
    }
    if d.smsDelivery == nil {
        return ErrSMSNotConfigured
    }
    return d.smsDelivery.Send(ctx, identifier, code)
}
```

This keeps the routing logic in one place and makes `GuestAccountService` agnostic to which transport is used.

---

## Tasks

### Phase 1: SMSConfig and ErrSMSNotConfigured (TDD)
- [ ] Write test: `TestSMSConfigFromEnv_Missing` — returns nil config and ErrSMSNotConfigured
- [ ] Write test: `TestSMSConfigFromEnv_Present` — returns populated config
- [ ] Run tests (should fail)
- [ ] Implement `SMSConfigFromEnv`
- [ ] Run tests (should pass)

### Phase 2: SMSOTPDelivery (TDD)
- [ ] Write test: `TestSMSOTPDelivery_Send_Success` — mock HTTP server returns 201, no error
- [ ] Write test: `TestSMSOTPDelivery_Send_ProviderError` — mock HTTP server returns 500, returns wrapped error
- [ ] Write test: `TestSMSOTPDelivery_Send_MessageFormat` — request body contains code but no URLs
- [ ] Run tests (should fail)
- [ ] Implement `SMSOTPDelivery`
- [ ] Run tests (should pass)

### Phase 3: Routing Delivery (TDD)
- [ ] Write test: `TestRoutingOTPDelivery_EmailRoute` — email identifier calls emailDelivery
- [ ] Write test: `TestRoutingOTPDelivery_PhoneRoute_SMSConfigured` — phone identifier calls smsDelivery
- [ ] Write test: `TestRoutingOTPDelivery_PhoneRoute_SMSNotConfigured` — returns ErrSMSNotConfigured
- [ ] Run tests (should fail)
- [ ] Implement `routingOTPDelivery` and update `NewOTPDelivery`
- [ ] Run tests (should pass)

### Phase 4: Wire at Startup
- [ ] Update `cmd/server/main.go` to call `SMSConfigFromEnv()` and pass to `NewOTPDelivery`
- [ ] Confirm app starts normally when `SMS_PROVIDER_URL` is unset (SMS simply not available)
- [ ] Run full test suite

---

## Testing Requirements

```go
func TestSMSOTPDelivery_Send_Success(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            t.Errorf("expected POST, got %s", r.Method)
        }
        r.ParseForm()
        if r.FormValue("To") != "+15559876543" {
            t.Errorf("unexpected To: %s", r.FormValue("To"))
        }
        if !strings.Contains(r.FormValue("Body"), "123456") {
            t.Errorf("expected code in body: %s", r.FormValue("Body"))
        }
        w.WriteHeader(http.StatusCreated)
    }))
    defer server.Close()

    cfg := &SMSConfig{
        ProviderURL: server.URL,
        FromNumber:  "+15550000000",
        AuthToken:   "test-token",
    }
    delivery := NewSMSOTPDelivery(cfg)
    err := delivery.Send(context.Background(), "+15559876543", "123456")
    if err != nil {
        t.Errorf("Send: %v", err)
    }
}

func TestSMSOTPDelivery_Send_MessageFormat(t *testing.T) {
    var capturedBody string
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        r.ParseForm()
        capturedBody = r.FormValue("Body")
        w.WriteHeader(http.StatusCreated)
    }))
    defer server.Close()

    cfg := &SMSConfig{ProviderURL: server.URL, FromNumber: "+15550000000", AuthToken: "tok"}
    delivery := NewSMSOTPDelivery(cfg)
    _ = delivery.Send(context.Background(), "+15551111111", "987654")

    if strings.Contains(capturedBody, "http") {
        t.Error("SMS body must not contain URLs")
    }
    if !strings.Contains(capturedBody, "987654") {
        t.Error("SMS body must contain the OTP code")
    }
}
```

---

## Dependencies

**Depends on:** Story 08 (guestauth package — `OTPDelivery` interface, `EmailOTPDelivery`)  
**Blocks:** Nothing (parallel to Stories 09–11; wired at startup only)

---

## Implementation Notes

### No SMS Library Dependency
All major SMS providers (Twilio, Vonage, Sinch, AWS SNS) offer a simple HTTP REST API with form-encoded POST. Using `net/http` directly avoids adding a provider-specific library that would need to be swapped if the operator uses a different provider.

### Provider URL Pattern
By including the Account SID in the `SMS_PROVIDER_URL`, operators can use the exact URL from their provider's documentation without the app needing to know the SID separately. HTTP Basic Auth uses the SID as username and the auth token as password, which is the Twilio-standard pattern.

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass: `go test -timeout 30s -race ./internal/guestauth/...`
- [ ] No real SMS provider calls in tests (all use mock HTTP server)
- [ ] App starts cleanly when `SMS_PROVIDER_URL` is unset
- [ ] `go vet ./...` clean
