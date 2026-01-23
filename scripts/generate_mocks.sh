#!/bin/bash
set -e

# Mock generation script for TinyRSVP
# This script generates all mocks for interfaces used in testing
# Mocks are organized into subdirectories to avoid import cycles

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Generating mocks for TinyRSVP...${NC}"

# Ensure mockgen is installed
if ! command -v mockgen &> /dev/null; then
    echo "mockgen not found. Installing..."
    go install go.uber.org/mock/mockgen@latest
fi

# Get GOPATH bin directory
MOCKGEN="$(go env GOPATH)/bin/mockgen"
MOCK_BASE="internal/testutil/mocks"

# Create mock subdirectories
mkdir -p "$MOCK_BASE/repositories"
mkdir -p "$MOCK_BASE/services"
mkdir -p "$MOCK_BASE/other"

echo -e "${GREEN}Generating repository mocks...${NC}"

# Repository mocks → mocks/repositories/
# These are safe to import from any package (no import cycles)

echo "  - EventRepository..."
$MOCKGEN -source=internal/db/repositories/event_repository.go \
    -destination=$MOCK_BASE/repositories/mock_event_repository.go \
    -package=repositories

echo "  - InviteRepository..."
$MOCKGEN -source=internal/db/repositories/invite_repository.go \
    -destination=$MOCK_BASE/repositories/mock_invite_repository.go \
    -package=repositories

echo "  - UserRepository..."
$MOCKGEN -source=internal/db/repositories/user_repository.go \
    -destination=$MOCK_BASE/repositories/mock_user_repository.go \
    -package=repositories

echo "  - RSVPRepository..."
$MOCKGEN -source=internal/db/repositories/rsvp_repository.go \
    -destination=$MOCK_BASE/repositories/mock_rsvp_repository.go \
    -package=repositories

echo "  - TemplateRepository..."
$MOCKGEN -source=internal/db/repositories/template_repository.go \
    -destination=$MOCK_BASE/repositories/mock_template_repository.go \
    -package=repositories

echo "  - AnswerRepository..."
$MOCKGEN -source=internal/db/repositories/answer_repository.go \
    -destination=$MOCK_BASE/repositories/mock_answer_repository.go \
    -package=repositories

echo "  - QuestionRepository..."
$MOCKGEN -source=internal/db/repositories/question_repository.go \
    -destination=$MOCK_BASE/repositories/mock_question_repository.go \
    -package=repositories

echo "  - ConfigRepository..."
$MOCKGEN -source=internal/db/repositories/config_repository.go \
    -destination=$MOCK_BASE/repositories/mock_config_repository.go \
    -package=repositories

echo "  - SessionRepository..."
$MOCKGEN -source=internal/db/repositories/session_repository.go \
    -destination=$MOCK_BASE/repositories/mock_session_repository.go \
    -package=repositories

echo "  - EmailQueueRepository..."
$MOCKGEN -source=internal/db/repositories/email_queue_repository.go \
    -destination=$MOCK_BASE/repositories/mock_email_queue_repository.go \
    -package=repositories

echo -e "${GREEN}Generating service mocks...${NC}"

# Service mocks → mocks/services/
# These import their service packages, so tests in those packages should use repository mocks instead

echo "  - EventService..."
$MOCKGEN -source=internal/events/service.go \
    -destination=$MOCK_BASE/services/mock_event_service.go \
    -package=services \
    -mock_names=Service=MockEventService

echo "  - InviteService..."
$MOCKGEN -source=internal/invites/service.go \
    -destination=$MOCK_BASE/services/mock_invite_service.go \
    -package=services

echo "  - RSVPService..."
$MOCKGEN -source=internal/rsvp/service.go \
    -destination=$MOCK_BASE/services/mock_rsvp_service.go \
    -package=services \
    -mock_names=Service=MockRSVPService \
    -exclude_interfaces=InviteService,InviteRepository

echo "  - TemplateService..."
$MOCKGEN -source=internal/templates/service.go \
    -destination=$MOCK_BASE/services/mock_template_service.go \
    -package=services \
    -mock_names=Service=MockTemplateService

echo "  - EmailService..."
$MOCKGEN -source=internal/email/service.go \
    -destination=$MOCK_BASE/services/mock_email_service.go \
    -package=services \
    -mock_names=Service=MockEmailService

echo -e "${GREEN}Generating other mocks...${NC}"

# Other mocks → mocks/other/
# Database, validators, storage, auth

echo "  - Database..."
$MOCKGEN -source=internal/db/db.go \
    -destination=$MOCK_BASE/other/mock_database.go \
    -package=other

echo "  - AuthorizationChecker..."
$MOCKGEN -source=internal/auth/permissions.go \
    -destination=$MOCK_BASE/other/mock_authorization.go \
    -package=other

echo "  - EventValidator..."
$MOCKGEN -source=internal/events/validator.go \
    -destination=$MOCK_BASE/other/mock_event_validator.go \
    -package=other \
    -mock_names=Validator=MockEventValidator

echo "  - TemplateValidator..."
$MOCKGEN -source=internal/templates/validator.go \
    -destination=$MOCK_BASE/other/mock_template_validator.go \
    -package=other \
    -mock_names=Validator=MockTemplateValidator

echo "  - StorageProvider..."
$MOCKGEN -source=internal/storage/provider.go \
    -destination=$MOCK_BASE/other/mock_storage_provider.go \
    -package=other

echo -e "${GREEN}Mock generation complete!${NC}"
echo -e "Generated mocks organized in:"
echo -e "  - $MOCK_BASE/repositories/ (10 mocks)"
echo -e "  - $MOCK_BASE/services/ (5 mocks)"
echo -e "  - $MOCK_BASE/other/ (5 mocks)"
