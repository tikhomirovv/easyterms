-- Core schema for EasyTerms MVP.
-- Entities: users, documents, document_sources, analysis_results, purchases, check_ledger.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Users: Telegram identity and check balance.
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_id     BIGINT NOT NULL,
    locale          VARCHAR(16) NOT NULL DEFAULT 'en',
    check_balance   INTEGER NOT NULL DEFAULT 0 CHECK (check_balance >= 0),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_telegram_id_unique UNIQUE (telegram_id)
);

CREATE INDEX idx_users_telegram_id ON users (telegram_id);

-- Documents: one paid check per document; original + clean text on the row.
CREATE TABLE documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status          TEXT NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft', 'ingested', 'paid')),
    check_consumed  BOOLEAN NOT NULL DEFAULT FALSE,
    original_text   TEXT,
    clean_text      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_documents_user_id_created_at ON documents (user_id, created_at DESC);

-- Document sources: fragments of input (paste, URL, later image).
CREATE TABLE document_sources (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id     UUID NOT NULL REFERENCES documents (id) ON DELETE CASCADE,
    kind            TEXT NOT NULL CHECK (kind IN ('text', 'url', 'image')),
    content         TEXT,
    source_url      TEXT,
    sequence        INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT document_sources_kind_content_check CHECK (
        (kind = 'text' AND content IS NOT NULL)
        OR (kind = 'url' AND source_url IS NOT NULL)
        OR (kind = 'image')
    )
);

CREATE INDEX idx_document_sources_document_id ON document_sources (document_id);

-- Analysis results: mode-specific payload in JSONB; one row per (document, type) for cache.
CREATE TABLE analysis_results (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id     UUID NOT NULL REFERENCES documents (id) ON DELETE CASCADE,
    analysis_type   TEXT NOT NULL,
    locale          VARCHAR(16) NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}',
    meta            JSONB NOT NULL DEFAULT '{}',
    cached          BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT analysis_results_document_type_unique UNIQUE (document_id, analysis_type)
);

CREATE INDEX idx_analysis_results_document_id ON analysis_results (document_id);

-- Purchases: payment attempts (stub, YooKassa, etc.).
CREATE TABLE purchases (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    provider_id         TEXT NOT NULL,
    package_id          TEXT NOT NULL,
    amount_cents        INTEGER NOT NULL CHECK (amount_cents >= 0),
    currency            CHAR(3) NOT NULL DEFAULT 'RUB',
    status              TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'completed', 'failed', 'cancelled')),
    external_payment_id TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT purchases_external_payment_id_unique UNIQUE (external_payment_id)
);

CREATE INDEX idx_purchases_user_id ON purchases (user_id);
CREATE INDEX idx_purchases_status ON purchases (status);

-- Check ledger: all balance movements (credits and debits).
CREATE TABLE check_ledger (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    document_id     UUID REFERENCES documents (id) ON DELETE SET NULL,
    purchase_id     UUID REFERENCES purchases (id) ON DELETE SET NULL,
    delta           INTEGER NOT NULL CHECK (delta <> 0),
    reason          TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_check_ledger_user_id_created_at ON check_ledger (user_id, created_at DESC);
