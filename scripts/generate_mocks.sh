#!/bin/bash
set -e

# Mock generation script for TinyRSVP
# This script generates all mocks for interfaces used in testing

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
MOCK_DIR="internal/testutil/mocks"

# Create mocks directory if it doesn't exist
mkdir -p "$MOCK_DIR"

echo -e "${GREEN}Generating Priority 1 mocks...${NC}"

# Story 06: Priority 1 Repository Mocks

# Database interface - foundation for all repositories
echo "  - Database interface..."
$MOCKGEN -source=internal/db/db.go \
    -destination=$MOCK_DIR/mock_database.go \
    -package=mocks

# EventRepository - 17 methods, widely used
echo "  - EventRepository..."
$MOCKGEN -source=internal/db/repositories/event_repository.go \
    -destination=$MOCK_DIR/mock_event_repository.go \
    -package=mocks

# InviteRepository - 13 methods
echo "  - InviteRepository..."
$MOCKGEN -source=internal/db/repositories/invite_repository.go \
    -destination=$MOCK_DIR/mock_invite_repository.go \
    -package=mocks

# UserRepository - 12 methods
echo "  - UserRepository..."
$MOCKGEN -source=internal/db/repositories/user_repository.go \
    -destination=$MOCK_DIR/mock_user_repository.go \
    -package=mocks

# AuthorizationChecker - permission testing
echo "  - AuthorizationChecker..."
$MOCKGEN -source=internal/auth/permissions.go \
    -destination=$MOCK_DIR/mock_authorization.go \
    -package=mocks

echo -e "${GREEN}Generating service mocks...${NC}"

# Story 07: Service Interface Mocks

# Event service - 8 methods
echo "  - EventService..."
$MOCKGEN -source=internal/events/service.go \
    -destination=$MOCK_DIR/mock_event_service.go \
    -package=mocks \
    -mock_names=Service=MockEventService

# Invite service - 16 methods
echo "  - InviteService..."
$MOCKGEN -source=internal/invites/service.go \
    -destination=$MOCK_DIR/mock_invite_service.go \
    -package=mocks

# RSVP service - 2 methods
echo "  - RSVPService..."
$MOCKGEN -source=internal/rsvp/service.go \
    -destination=$MOCK_DIR/mock_rsvp_service.go \
    -package=mocks \
    -mock_names=Service=MockRSVPService \
    -exclude_interfaces=InviteService,InviteRepository

# Template service - 11 methods
echo "  - TemplateService..."
$MOCKGEN -source=internal/templates/service.go \
    -destination=$MOCK_DIR/mock_template_service.go \
    -package=mocks \
    -mock_names=Service=MockTemplateService

# Email service - 1 method
echo "  - EmailService..."
$MOCKGEN -source=internal/email/service.go \
    -destination=$MOCK_DIR/mock_email_service.go \
    -package=mocks \
    -mock_names=Service=MockEmailService

echo -e "${GREEN}Generating remaining repository mocks...${NC}"

# Story 08: Remaining Repository Mocks

# RSVPRepository
echo "  - RSVPRepository..."
$MOCKGEN -source=internal/db/repositories/rsvp_repository.go \
    -destination=$MOCK_DIR/mock_rsvp_repository.go \
    -package=mocks

# TemplateRepository
echo "  - TemplateRepository..."
$MOCKGEN -source=internal/db/repositories/template_repository.go \
    -destination=$MOCK_DIR/mock_template_repository.go \
    -package=mocks

# AnswerRepository
echo "  - AnswerRepository..."
$MOCKGEN -source=internal/db/repositories/answer_repository.go \
    -destination=$MOCK_DIR/mock_answer_repository.go \
    -package=mocks

# QuestionRepository
echo "  - QuestionRepository..."
$MOCKGEN -source=internal/db/repositories/question_repository.go \
    -destination=$MOCK_DIR/mock_question_repository.go \
    -package=mocks

# ConfigRepository
echo "  - ConfigRepository..."
$MOCKGEN -source=internal/db/repositories/config_repository.go \
    -destination=$MOCK_DIR/mock_config_repository.go \
    -package=mocks

# SessionRepository
echo "  - SessionRepository..."
$MOCKGEN -source=internal/db/repositories/session_repository.go \
    -destination=$MOCK_DIR/mock_session_repository.go \
    -package=mocks

# EmailQueueRepository
echo "  - EmailQueueRepository..."
$MOCKGEN -source=internal/db/repositories/email_queue_repository.go \
    -destination=$MOCK_DIR/mock_email_queue_repository.go \
    -package=mocks

echo -e "${GREEN}Generating validator and utility mocks...${NC}"

# Validators
echo "  - EventValidator..."
$MOCKGEN -source=internal/events/validator.go \
    -destination=$MOCK_DIR/mock_event_validator.go \
    -package=mocks \
    -mock_names=Validator=MockEventValidator

echo "  - TemplateValidator..."
$MOCKGEN -source=internal/templates/validator.go \
    -destination=$MOCK_DIR/mock_template_validator.go \
    -package=mocks \
    -mock_names=Validator=MockTemplateValidator

# Storage Provider
echo "  - StorageProvider..."
$MOCKGEN -source=internal/storage/provider.go \
    -destination=$MOCK_DIR/mock_storage_provider.go \
    -package=mocks

echo -e "${GREEN}Mock generation complete!${NC}"
echo -e "Generated mocks are in: $MOCK_DIR"
