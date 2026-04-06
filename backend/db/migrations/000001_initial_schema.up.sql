-- Initial schema with hybrid IDs: BIGSERIAL for internal use + UUID for public API.
-- All foreign keys reference the internal BIGSERIAL id for fast joins.
-- The guid (UUID) column is exposed in API responses and URL params.

CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL PRIMARY KEY,
    guid  UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    email      VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role       VARCHAR(32) NOT NULL DEFAULT 'admin',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sites (
    id                 BIGSERIAL PRIMARY KEY,
    guid          UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    name               VARCHAR(255) NOT NULL,
    base_url           VARCHAR(512) NOT NULL,
    is_active          BOOLEAN NOT NULL DEFAULT true,
    daily_hit_limit    INTEGER NOT NULL DEFAULT 100,
    hit_delay_ms       INTEGER NOT NULL DEFAULT 2000,
    reparse_enabled    BOOLEAN NOT NULL DEFAULT false,
    reparse_after_days INTEGER NOT NULL DEFAULT 30,
    user_agent         VARCHAR(255) NOT NULL DEFAULT '',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- learn_categories defined early so site_categories can reference it
CREATE TABLE IF NOT EXISTS learn_categories (
    id         BIGSERIAL PRIMARY KEY,
    guid  UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    name       VARCHAR(128) NOT NULL,
    color      VARCHAR(16) NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS site_categories (
    id                BIGSERIAL PRIMARY KEY,
    guid         UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    site_id           BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name              VARCHAR(255) NOT NULL,
    start_url         VARCHAR(512) NOT NULL DEFAULT '',
    url_pattern       VARCHAR(512) NOT NULL DEFAULT '',
    selector_title    VARCHAR(255) NOT NULL DEFAULT '',
    selector_body     VARCHAR(255) NOT NULL DEFAULT '',
    source_language   VARCHAR(8) NOT NULL DEFAULT 'en',
    is_active         BOOLEAN NOT NULL DEFAULT true,
    learn_category_id BIGINT REFERENCES learn_categories(id) ON DELETE SET NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_site_categories_site_id ON site_categories(site_id);

CREATE TABLE IF NOT EXISTS parse_jobs (
    id               BIGSERIAL PRIMARY KEY,
    guid        UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    site_id          BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    site_category_id BIGINT REFERENCES site_categories(id) ON DELETE SET NULL,
    url              VARCHAR(1024) NOT NULL,
    depth            INTEGER NOT NULL DEFAULT 0,
    status           VARCHAR(32) NOT NULL DEFAULT 'pending',
    scheduled_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at       TIMESTAMPTZ,
    finished_at      TIMESTAMPTZ,
    attempt_count    INTEGER NOT NULL DEFAULT 0,
    last_error       VARCHAR(1024) NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_parse_jobs_site_id ON parse_jobs(site_id);
CREATE INDEX IF NOT EXISTS idx_parse_jobs_status ON parse_jobs(status);
CREATE INDEX IF NOT EXISTS idx_parse_jobs_scheduled_at ON parse_jobs(scheduled_at);

CREATE TABLE IF NOT EXISTS parsed_pages (
    id               BIGSERIAL PRIMARY KEY,
    guid        UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    site_id          BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    site_category_id BIGINT REFERENCES site_categories(id) ON DELETE SET NULL,
    url              VARCHAR(1024) NOT NULL,
    url_hash         VARCHAR(64) NOT NULL,
    content_hash     VARCHAR(64) NOT NULL DEFAULT '',
    title            VARCHAR(512) NOT NULL DEFAULT '',
    word_count       INTEGER NOT NULL DEFAULT 0,
    parsed_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_site_url ON parsed_pages(site_id, url_hash);
CREATE INDEX IF NOT EXISTS idx_parsed_pages_parsed_at ON parsed_pages(parsed_at);

CREATE TABLE IF NOT EXISTS words (
    id             BIGSERIAL PRIMARY KEY,
    guid      UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    lemma          VARCHAR(128) NOT NULL,
    language       VARCHAR(8) NOT NULL DEFAULT 'en',
    pos            VARCHAR(32) NOT NULL DEFAULT '',
    frequency_rank INTEGER,
    ipa            VARCHAR(128) NOT NULL DEFAULT '',
    definition     VARCHAR(2048) NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_lemma_lang ON words(lemma, language);
CREATE INDEX IF NOT EXISTS idx_words_frequency_rank ON words(frequency_rank);

CREATE TABLE IF NOT EXISTS word_occurrences (
    id              BIGSERIAL PRIMARY KEY,
    guid       UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    word_id         BIGINT NOT NULL REFERENCES words(id) ON DELETE CASCADE,
    parsed_page_id  BIGINT NOT NULL REFERENCES parsed_pages(id) ON DELETE CASCADE,
    count           INTEGER NOT NULL DEFAULT 1,
    sample_sentence VARCHAR(1024) NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_word_page ON word_occurrences(word_id, parsed_page_id);

CREATE TABLE IF NOT EXISTS word_translations (
    id              BIGSERIAL PRIMARY KEY,
    guid       UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    word_id         BIGINT NOT NULL REFERENCES words(id) ON DELETE CASCADE,
    target_language VARCHAR(8) NOT NULL,
    translation     VARCHAR(512) NOT NULL,
    source          VARCHAR(32) NOT NULL DEFAULT '',
    is_primary      BOOLEAN NOT NULL DEFAULT false,
    fetched_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_word_lang_trans ON word_translations(word_id, target_language, translation, source);

CREATE TABLE IF NOT EXISTS word_relations (
    id            BIGSERIAL PRIMARY KEY,
    guid     UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    word_id       BIGINT NOT NULL REFERENCES words(id) ON DELETE CASCADE,
    related_text  VARCHAR(128) NOT NULL,
    relation_type VARCHAR(16) NOT NULL,
    source        VARCHAR(32) NOT NULL DEFAULT '',
    fetched_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_word_rel ON word_relations(word_id, related_text, relation_type, source);

CREATE TABLE IF NOT EXISTS review_items (
    id                 BIGSERIAL PRIMARY KEY,
    guid          UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    word_id            BIGINT NOT NULL UNIQUE REFERENCES words(id) ON DELETE CASCADE,
    first_seen_page_id BIGINT REFERENCES parsed_pages(id) ON DELETE SET NULL,
    site_category_id   BIGINT REFERENCES site_categories(id) ON DELETE SET NULL,
    status             VARCHAR(16) NOT NULL DEFAULT 'pending',
    reviewed_at        TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_review_items_status ON review_items(status);

CREATE TABLE IF NOT EXISTS learn_items (
    id               BIGSERIAL PRIMARY KEY,
    guid        UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    word_id          BIGINT NOT NULL UNIQUE REFERENCES words(id) ON DELETE CASCADE,
    status           VARCHAR(16) NOT NULL DEFAULT 'active',
    mastery_level    INTEGER NOT NULL DEFAULT 0,
    last_reviewed_at TIMESTAMPTZ,
    archived_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_learn_items_status ON learn_items(status);

CREATE TABLE IF NOT EXISTS learn_item_categories (
    learn_item_id     BIGINT NOT NULL REFERENCES learn_items(id) ON DELETE CASCADE,
    learn_category_id BIGINT NOT NULL REFERENCES learn_categories(id) ON DELETE CASCADE,
    PRIMARY KEY (learn_item_id, learn_category_id)
);
CREATE INDEX IF NOT EXISTS idx_learn_item_categories_category ON learn_item_categories(learn_category_id);

CREATE TABLE IF NOT EXISTS app_settings (
    key        VARCHAR(64) PRIMARY KEY,
    value      VARCHAR(2048) NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Geo lookup tables (no public UUID needed — internal only)
CREATE TABLE IF NOT EXISTS countries (
    id   BIGSERIAL PRIMARY KEY,
    name VARCHAR(128) NOT NULL UNIQUE,
    code VARCHAR(8) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_countries_code ON countries(code);

CREATE TABLE IF NOT EXISTS states (
    id   BIGSERIAL PRIMARY KEY,
    name VARCHAR(128) NOT NULL UNIQUE,
    code VARCHAR(8) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_states_code ON states(code);

CREATE TABLE IF NOT EXISTS cities (
    id   BIGSERIAL PRIMARY KEY,
    name VARCHAR(128) NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_cities_name ON cities(name);
