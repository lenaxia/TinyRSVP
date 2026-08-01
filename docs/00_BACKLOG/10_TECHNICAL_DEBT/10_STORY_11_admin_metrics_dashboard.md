# Epic 10: Technical Debt & Improvements
## Story 04: Admin Metrics Dashboard

### User Story
As an admin, I want to see useful system metrics in a dashboard format instead of raw Prometheus metrics so that I can monitor system health without technical knowledge.

### Problem
The admin dashboard has a "Metrics" quick action that links to `/metrics`, which displays raw Prometheus metrics endpoint. This is not useful for non-technical admins and should instead show a formatted dashboard with key metrics.

### Acceptance Criteria
- [ ] Create `/admin/metrics` route for admin metrics dashboard
- [ ] Create metrics dashboard handler with admin authorization
- [ ] Create metrics dashboard template with formatted metrics display
- [ ] Display key metrics in human-readable format:
  - HTTP request rates and response times
  - Error rates
  - Email queue status
  - Database connection pool stats
  - Active sessions count
- [ ] Keep raw `/metrics` endpoint for Prometheus scraping
- [ ] Update admin dashboard link to point to `/admin/metrics`
- [ ] Add tests for metrics dashboard

### Technical Notes
- Raw `/metrics` endpoint should remain unchanged for Prometheus integration
- New `/admin/metrics` route should parse and format metrics for display
- Consider using charts/graphs for time-series data
- May need to store historical metrics data or query from Prometheus
- Initial implementation can show current snapshot of metrics

### Alternative Approaches
1. Create admin-friendly metrics dashboard (recommended)
2. Remove metrics link from admin dashboard entirely
3. Link to external monitoring tool (Grafana) if available

### Status
- Status: ✅ Complete (verified 2026-08-01: `/admin/metrics` + `AdminMetricsStats`; worklog 0175a)
- Priority: Low
- Assigned: Unassigned
- Created: 2026-01-10
