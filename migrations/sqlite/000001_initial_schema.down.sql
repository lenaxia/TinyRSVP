-- TinyRSVP Initial Schema Rollback
-- Drops all 11 core tables in reverse dependency order
-- Migration: 000001_initial_schema

-- Drop tables in reverse order of creation to respect foreign key constraints

DROP TABLE IF EXISTS config;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS email_queue;
DROP TABLE IF EXISTS rsvp_answers;
DROP TABLE IF EXISTS preference_questions;
DROP TABLE IF EXISTS rsvps;
DROP TABLE IF EXISTS invites;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
