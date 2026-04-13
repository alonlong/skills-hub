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
