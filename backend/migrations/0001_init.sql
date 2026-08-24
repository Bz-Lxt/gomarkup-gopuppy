CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    nickname TEXT NOT NULL,
    avatar_url TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS families (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS family_members (
    family_id UUID NOT NULL REFERENCES families(id),
    user_id UUID NOT NULL REFERENCES users(id),
    role TEXT NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (family_id, user_id)
);

CREATE TABLE IF NOT EXISTS family_invites (
    id UUID PRIMARY KEY,
    family_id UUID NOT NULL REFERENCES families(id),
    code TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_by UUID REFERENCES users(id),
    used_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS pets (
    id UUID PRIMARY KEY,
    family_id UUID NOT NULL REFERENCES families(id),
    name TEXT NOT NULL,
    species TEXT NOT NULL,
    breed TEXT NOT NULL DEFAULT '',
    gender TEXT NOT NULL,
    birthday DATE NOT NULL,
    avatar_key TEXT NOT NULL DEFAULT '',
    neutered BOOLEAN NOT NULL DEFAULT FALSE,
    chip_no TEXT NOT NULL DEFAULT '',
    weight_min NUMERIC(5,2),
    weight_max NUMERIC(5,2),
    note TEXT NOT NULL DEFAULT '',
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_pets_family ON pets(family_id);

CREATE TABLE IF NOT EXISTS daily_checkins (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets(id),
    checkin_date DATE NOT NULL,
    type TEXT NOT NULL,
    slot TEXT NOT NULL,
    done_by UUID NOT NULL REFERENCES users(id),
    done_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    UNIQUE (pet_id, checkin_date, type, slot)
);

CREATE TABLE IF NOT EXISTS health_events (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets(id),
    category TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    occurred_at TIMESTAMPTZ NOT NULL,
    clinic TEXT NOT NULL DEFAULT '',
    severity TEXT NOT NULL DEFAULT '',
    treated BOOLEAN NOT NULL DEFAULT FALSE,
    amount_cents BIGINT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_events_pet ON health_events(pet_id, occurred_at DESC);

CREATE TABLE IF NOT EXISTS reminder_rules (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets(id),
    kind TEXT NOT NULL,
    title TEXT NOT NULL,
    cycle_days INT NOT NULL CHECK (cycle_days > 0),
    last_done_at TIMESTAMPTZ NOT NULL,
    next_due_at DATE NOT NULL,
    advance_days INT NOT NULL DEFAULT 3,
    channels TEXT[] NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS notification_logs (
    id UUID PRIMARY KEY,
    rule_id UUID NOT NULL REFERENCES reminder_rules(id),
    pet_id UUID NOT NULL REFERENCES pets(id),
    due_date DATE NOT NULL,
    channel TEXT NOT NULL,
    kind TEXT NOT NULL,
    status TEXT NOT NULL,
    attempt INT NOT NULL DEFAULT 0,
    error TEXT NOT NULL DEFAULT '',
    scheduled_at TIMESTAMPTZ NOT NULL,
    sent_at TIMESTAMPTZ,
    UNIQUE (rule_id, due_date, channel, kind)
);

CREATE TABLE IF NOT EXISTS weight_records (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets(id),
    weight_kg NUMERIC(5,2) NOT NULL,
    measured_at TIMESTAMPTZ NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    created_by UUID NOT NULL REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS expenses (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets(id),
    category TEXT NOT NULL,
    amount_cents BIGINT NOT NULL CHECK (amount_cents >= 0),
    spent_at TIMESTAMPTZ NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    created_by UUID NOT NULL REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS media_files (
    id UUID PRIMARY KEY,
    family_id UUID NOT NULL REFERENCES families(id),
    pet_id UUID NOT NULL REFERENCES pets(id),
    kind TEXT NOT NULL,
    storage_driver TEXT NOT NULL,
    object_key TEXT NOT NULL,
    filename TEXT NOT NULL,
    mime TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    sha256 TEXT NOT NULL,
    uploaded_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_media_family_sha ON media_files(family_id, sha256);

CREATE TABLE IF NOT EXISTS event_attachments (
    event_id UUID NOT NULL REFERENCES health_events(id) ON DELETE CASCADE,
    media_id UUID NOT NULL REFERENCES media_files(id) ON DELETE CASCADE,
    PRIMARY KEY (event_id, media_id)
);

CREATE TABLE IF NOT EXISTS mock_deliveries (
    id UUID PRIMARY KEY,
    channel TEXT NOT NULL,
    payload TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
