package handlers

const RouterDocumentation = `
TinyRSVP HTTP Router Documentation
===================================

Route Organization
------------------

PUBLIC ROUTES (No Authentication Required):
- GET  /health                          - Health check endpoint
- GET  /ready                           - Readiness check endpoint
- GET  /login                           - OIDC login initiation
- GET  /auth/login                      - Alternative login path
- GET  /auth/callback                   - OIDC callback handler
- POST /logout                          - Logout handler
- POST /auth/logout                     - Alternative logout path
- GET  /rsvp/{token}                    - Guest RSVP page (token-based access)
- POST /rsvp/{token}                    - Submit RSVP response
- PUT  /rsvp/{token}                    - Update existing RSVP
- GET  /rsvp/{token}/confirmation       - RSVP confirmation page
- GET  /static/*                        - Static file serving (CSS, JS, images)
- GET  /assets/*                        - Dynamic asset serving (event images)

AUTHENTICATED ROUTES (Requires Auth Middleware):
- GET  /                                - Dashboard (event statistics and recent activity)

Web UI Event Management (HTML Forms):
- GET  /events                          - List events page
- GET  /events/new                      - New event form
- POST /events                          - Create event from form
- GET  /events/{id}                     - View event details page
- GET  /events/{id}/edit                - Edit event form
- POST /events/{id}                     - Update event from form
- POST /events/{id}/publish             - Publish event action
- POST /events/{id}/cancel              - Cancel event action (requires reason)
- POST /events/{id}/delete              - Delete event action

All routes under /api/* require authentication via AuthMiddleware.RequireAuth

API Event Management (JSON):
- GET    /api/events                    - List all events
- POST   /api/events                    - Create new event
- GET    /api/events/{id}               - Get event details
- PUT    /api/events/{id}               - Update event
- DELETE /api/events/{id}               - Delete event
- GET    /api/events/{id}/rsvp-summary  - Get RSVP summary for event

Event Questions:
- Routes registered via QuestionHandlers.RegisterRoutes()
- Manages custom questions for events

Invite Management:
- GET  /api/events/{eventId}/invites         - List invites for event
- POST /api/events/{eventId}/invites         - Create single invite
- POST /api/events/{eventId}/invites/import  - Bulk import invites from CSV
- POST /api/events/{eventId}/invites/manual  - Create manual invite
- POST /api/invites/{inviteId}/revoke        - Revoke an invite
- POST /api/invites/{inviteId}/regenerate    - Regenerate invite token

Image Management:
- Routes registered via ImageHandlers.RegisterRoutes()
- Handles event image uploads and serving

Template Management:
- Routes registered via TemplateHandlers.RegisterRoutes()
- Serves web UI templates (dashboard, event forms, etc.)

ADMIN ROUTES (Requires Auth + Admin Middleware):
All routes require both AuthMiddleware.RequireAuth AND AuthMiddleware.RequireAdmin

User Management:
- GET    /api/users           - List all users
- GET    /api/users/{userId}  - Get user details
- PATCH  /api/users/{userId}  - Update user role
- DELETE /api/users/{userId}  - Delete user

System Operations:
- POST /api/invites/cleanup   - Cleanup expired invites
- GET  /api/email/health      - Email service health check

Route Parameters
----------------

Path Parameters:
- {id}       - Event ID (int64, positive)
- {eventId}  - Event ID for nested resources (int64, positive)
- {inviteId} - Invite ID (int64, positive)
- {userId}   - User ID (string)
- {token}    - Invite token (string, alphanumeric with dashes)

Parameter Extraction:
Use GetEventIDFromRequest(r) for extracting and validating event IDs
Use GetTokenFromRequest(r) for extracting and validating tokens

Error Handling
--------------

404 Not Found:
- API requests (Accept: application/json or /api/* path): JSON error response
- Web requests: HTML 404 page

405 Method Not Allowed:
- API requests: JSON error response
- Web requests: HTML 405 page

Middleware Chain
----------------

Global Middleware (applied to all routes):
1. middleware.RequestID  - Adds unique request ID
2. middleware.RealIP     - Extracts real client IP
3. middleware.Logger     - Logs all requests
4. middleware.Recoverer  - Recovers from panics

API Middleware (applied to /api/* routes):
1. AuthMiddleware.RequireAuth - Validates authentication

Admin Middleware (applied to admin routes):
1. AuthMiddleware.RequireAuth  - Validates authentication
2. AuthMiddleware.RequireAdmin - Validates admin role

Request/Response Formats
------------------------

API Routes:
- Request: application/json
- Response: application/json
- Errors: JSON with "error" field

Web Routes:
- Request: application/x-www-form-urlencoded or multipart/form-data
- Response: text/html
- Errors: HTML error pages

RSVP Routes:
- Request: application/x-www-form-urlencoded
- Response: text/html (pages) or redirects
- No authentication required (token-based access)

Debugging
---------

Use router.ListRoutes() to enumerate all registered routes for debugging.
Returns []RouteInfo with Method and Pattern for each route.
`
