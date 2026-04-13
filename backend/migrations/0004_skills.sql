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
