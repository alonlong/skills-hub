-- Full PostgreSQL schema for SkillHub (single source of truth).
-- Used by Docker initdb (docker-compose) and any migrate tooling that applies this directory.
-- Label tables at the end use IF NOT EXISTS so that block can be applied alone to DBs missing labels.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name VARCHAR(128) NOT NULL,
    email VARCHAR(256),
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE namespaces (
    id UUID PRIMARY KEY,
    slug VARCHAR(64) NOT NULL UNIQUE,
    display_name VARCHAR(128) NOT NULL,
    type VARCHAR(16) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE namespace_members (
    id UUID PRIMARY KEY,
    namespace_id UUID NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(16) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(namespace_id, user_id)
);

CREATE TABLE skills (
    id UUID PRIMARY KEY,
    namespace_id UUID NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    slug VARCHAR(128) NOT NULL,
    display_name VARCHAR(256) NOT NULL,
    summary VARCHAR(512) NOT NULL DEFAULT '',
    owner_id UUID NOT NULL REFERENCES users(id),
    visibility VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    latest_version_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(namespace_id, slug)
);

CREATE TABLE skill_versions (
    id UUID PRIMARY KEY,
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    version VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    manifest_json JSONB NOT NULL,
    parsed_metadata_json JSONB NOT NULL,
    storage_path TEXT NOT NULL,
    submitted_by UUID NOT NULL REFERENCES users(id),
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(skill_id, version)
);

ALTER TABLE skills
    ADD CONSTRAINT fk_skills_latest_version
    FOREIGN KEY (latest_version_id) REFERENCES skill_versions(id);

CREATE TABLE skill_tags (
    id UUID PRIMARY KEY,
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    tag_name VARCHAR(64) NOT NULL,
    target_version_id UUID NOT NULL REFERENCES skill_versions(id) ON DELETE CASCADE,
    UNIQUE(skill_id, tag_name)
);

CREATE TABLE review_tasks (
    id UUID PRIMARY KEY,
    skill_version_id UUID NOT NULL REFERENCES skill_versions(id) ON DELETE CASCADE,
    namespace_id UUID NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL,
    decision VARCHAR(32),
    reviewer_id UUID REFERENCES users(id),
    comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ
);

CREATE TABLE search_documents (
    skill_version_id UUID PRIMARY KEY REFERENCES skill_versions(id) ON DELETE CASCADE,
    namespace_slug VARCHAR(64) NOT NULL,
    skill_slug VARCHAR(128) NOT NULL,
    title TEXT NOT NULL,
    summary TEXT NOT NULL,
    content TEXT NOT NULL,
    visibility VARCHAR(32) NOT NULL,
    document tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(summary, '')), 'B') ||
        setweight(to_tsvector('simple', coalesce(content, '')), 'C')
    ) STORED
);

CREATE INDEX idx_search_documents_fts ON search_documents USING GIN (document);

-- Label definitions for filters and skill tagging (admin-managed).

CREATE TABLE IF NOT EXISTS label_definitions (
    slug VARCHAR(64) PRIMARY KEY,
    type VARCHAR(32) NOT NULL,
    visible_in_filter BOOLEAN NOT NULL DEFAULT true,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS label_translations (
    label_slug VARCHAR(64) NOT NULL REFERENCES label_definitions (slug) ON DELETE CASCADE,
    locale VARCHAR(32) NOT NULL,
    display_name VARCHAR(256) NOT NULL,
    PRIMARY KEY (label_slug, locale)
);

CREATE INDEX IF NOT EXISTS idx_label_definitions_sort ON label_definitions (sort_order, slug);
