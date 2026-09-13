-- Core schema for EasyTerms (SQLite).
-- Entities: users, documents, document_sources, analysis_results.

PRAGMA foreign_keys = ON;

CREATE TABLE users (
    id              TEXT PRIMARY KEY,
    telegram_id     INTEGER NOT NULL,
    locale          TEXT NOT NULL DEFAULT 'en',
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now')),
    CONSTRAINT users_telegram_id_unique UNIQUE (telegram_id)
);

CREATE INDEX idx_users_telegram_id ON users (telegram_id);

CREATE TABLE documents (
    id              TEXT PRIMARY KEY,
    user_id         TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status          TEXT NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft', 'ingested')),
    original_text   TEXT,
    clean_text      TEXT,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_documents_user_id_created_at ON documents (user_id, created_at DESC);

CREATE TABLE document_sources (
    id              TEXT PRIMARY KEY,
    document_id     TEXT NOT NULL REFERENCES documents (id) ON DELETE CASCADE,
    kind            TEXT NOT NULL CHECK (kind IN ('text', 'url', 'image')),
    content         TEXT,
    source_url      TEXT,
    sequence        INTEGER NOT NULL DEFAULT 0,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    CONSTRAINT document_sources_kind_content_check CHECK (
        (kind = 'text' AND content IS NOT NULL)
        OR (kind = 'url' AND source_url IS NOT NULL)
        OR (kind = 'image')
    )
);

CREATE INDEX idx_document_sources_document_id ON document_sources (document_id);

CREATE TABLE analysis_results (
    id              TEXT PRIMARY KEY,
    document_id     TEXT NOT NULL REFERENCES documents (id) ON DELETE CASCADE,
    analysis_type   TEXT NOT NULL,
    locale          TEXT NOT NULL,
    payload         TEXT NOT NULL DEFAULT '{}',
    meta            TEXT NOT NULL DEFAULT '{}',
    cached          INTEGER NOT NULL DEFAULT 0,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    CONSTRAINT analysis_results_document_type_unique UNIQUE (document_id, analysis_type)
);

CREATE INDEX idx_analysis_results_document_id ON analysis_results (document_id);
