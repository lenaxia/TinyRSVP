-- TinyRSVP Initial Schema
-- Creates all 11 core tables with constraints and indexes
-- Migration: 000001_initial_schema

-- Table 1: users
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    name TEXT,
    role TEXT NOT NULL DEFAULT 'event_manager',
    oidc_subject TEXT UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    CHECK (role IN ('admin', 'event_manager'))
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_oidc_subject ON users(oidc_subject);
CREATE INDEX idx_users_role ON users(role);

-- Table 2: sessions
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    last_accessed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Table 3: templates
CREATE TABLE templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    html_content TEXT NOT NULL,
    text_content TEXT,
    css_content TEXT,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (type IN ('invite_email', 'rsvp_page', 'confirmation_page'))
);

CREATE INDEX idx_templates_type ON templates(type);
CREATE INDEX idx_templates_is_default ON templates(is_default);
CREATE INDEX idx_templates_created_by ON templates(created_by);

-- Table 4: events
CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    location TEXT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    timezone TEXT NOT NULL,
    rsvp_deadline TIMESTAMP,
    max_plus_ones INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'draft',
    template_id INTEGER REFERENCES templates(id) ON DELETE SET NULL,
    created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    ics_sequence INTEGER NOT NULL DEFAULT 0,
    CHECK (status IN ('draft', 'published', 'cancelled', 'archived')),
    CHECK (max_plus_ones >= 0 AND max_plus_ones <= 10),
    CHECK (end_time IS NULL OR end_time > start_time),
    CHECK (rsvp_deadline IS NULL OR rsvp_deadline < start_time)
);

CREATE INDEX idx_events_created_by ON events(created_by);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_start_time ON events(start_time);

-- Table 5: invites
CREATE TABLE invites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name TEXT,
    email TEXT,
    token_hash TEXT NOT NULL UNIQUE,
    max_plus_ones INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    sent_at TIMESTAMP,
    viewed_at TIMESTAMP,
    unsubscribed BOOLEAN NOT NULL DEFAULT FALSE,
    email_invalid BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    CHECK (status IN ('draft', 'sent', 'viewed', 'responded', 'revoked')),
    CHECK (max_plus_ones >= 0 AND max_plus_ones <= 10)
);

CREATE INDEX idx_invites_event_id ON invites(event_id);
CREATE INDEX idx_invites_token_hash ON invites(token_hash);
CREATE INDEX idx_invites_email ON invites(email);
CREATE INDEX idx_invites_status ON invites(status);
CREATE INDEX idx_invites_expires_at ON invites(expires_at);

-- Table 6: rsvps
CREATE TABLE rsvps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invite_id INTEGER NOT NULL UNIQUE REFERENCES invites(id) ON DELETE CASCADE,
    response TEXT NOT NULL,
    plus_ones INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (response IN ('yes', 'no', 'maybe')),
    CHECK (plus_ones >= 0)
);

CREATE INDEX idx_rsvps_invite_id ON rsvps(invite_id);
CREATE INDEX idx_rsvps_response ON rsvps(response);

-- Table 7: preference_questions
CREATE TABLE preference_questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    question_type TEXT NOT NULL,
    options JSON,
    required BOOLEAN NOT NULL DEFAULT FALSE,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (question_type IN ('text', 'select', 'boolean'))
);

CREATE INDEX idx_questions_event_id ON preference_questions(event_id);
CREATE INDEX idx_questions_display_order ON preference_questions(event_id, display_order);

-- Table 8: rsvp_answers
CREATE TABLE rsvp_answers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rsvp_id INTEGER NOT NULL REFERENCES rsvps(id) ON DELETE CASCADE,
    question_id INTEGER NOT NULL REFERENCES preference_questions(id) ON DELETE CASCADE,
    answer_text TEXT,
    answer_option TEXT,
    answer_boolean BOOLEAN,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(rsvp_id, question_id)
);

CREATE INDEX idx_answers_rsvp_id ON rsvp_answers(rsvp_id);
CREATE INDEX idx_answers_question_id ON rsvp_answers(question_id);

-- Table 9: email_queue
CREATE TABLE email_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    to_email TEXT NOT NULL,
    to_name TEXT,
    subject TEXT NOT NULL,
    body_text TEXT NOT NULL,
    body_html TEXT,
    attachments JSON,
    status TEXT NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 4,
    last_attempt_at TIMESTAMP,
    last_error TEXT,
    scheduled_for TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (status IN ('pending', 'sending', 'sent', 'failed', 'cancelled'))
);

CREATE INDEX idx_email_queue_status_scheduled ON email_queue(status, scheduled_for);
CREATE INDEX idx_email_queue_status ON email_queue(status);

-- Table 10: audit_log
CREATE TABLE audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id INTEGER,
    details JSON,
    ip_address TEXT,
    user_agent TEXT
);

CREATE INDEX idx_audit_log_timestamp ON audit_log(timestamp);
CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_log_resource ON audit_log(resource_type, resource_id);

-- Table 11: config
CREATE TABLE config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
