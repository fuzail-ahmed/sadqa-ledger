-- Initial schema: all six application tables, per docs/SCHEMA.md §3.
-- Table order follows FK dependencies (admins/members/group_settings first).

CREATE TABLE admins (
    id              INTEGER PRIMARY KEY,
    username        TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    display_name    TEXT NOT NULL,
    language_pref   TEXT NOT NULL DEFAULT 'en',
    is_active       INTEGER NOT NULL DEFAULT 1,
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE members (
    id              INTEGER PRIMARY KEY,
    name            TEXT NOT NULL,
    is_active       INTEGER NOT NULL DEFAULT 1,
    notes           TEXT,
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE INDEX idx_members_active_name ON members (is_active, name);

CREATE TABLE group_settings (
    id                       INTEGER PRIMARY KEY CHECK (id = 1),
    group_name               TEXT NOT NULL,
    currency_code            TEXT NOT NULL DEFAULT 'INR',
    currency_symbol          TEXT NOT NULL DEFAULT '₹',
    show_names_publicly      INTEGER NOT NULL DEFAULT 0,
    public_token             TEXT NOT NULL UNIQUE,
    quick_amounts_minor      TEXT NOT NULL DEFAULT '[20000,50000,100000,200000]',
    default_public_language  TEXT NOT NULL DEFAULT 'en',
    privacy_policy_url       TEXT,
    updated_at               TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE contributions (
    id                    INTEGER PRIMARY KEY,
    member_id             INTEGER NOT NULL REFERENCES members(id),
    amount_minor          INTEGER NOT NULL CHECK (amount_minor > 0),
    contribution_month    TEXT NOT NULL,
    paid_on               TEXT NOT NULL,
    recorded_by_admin_id  INTEGER NOT NULL REFERENCES admins(id),
    deleted_at            TEXT,
    deleted_by_admin_id   INTEGER REFERENCES admins(id),
    created_at            TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at            TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE INDEX idx_contributions_member_month ON contributions (member_id, contribution_month);
CREATE INDEX idx_contributions_month ON contributions (contribution_month) WHERE deleted_at IS NULL;

CREATE TABLE expenses (
    id                    INTEGER PRIMARY KEY,
    description           TEXT NOT NULL,
    amount_minor          INTEGER NOT NULL CHECK (amount_minor > 0),
    expense_date          TEXT NOT NULL,
    receipt_photo_path    TEXT,
    recorded_by_admin_id  INTEGER NOT NULL REFERENCES admins(id),
    deleted_at            TEXT,
    deleted_by_admin_id   INTEGER REFERENCES admins(id),
    created_at            TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at            TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE INDEX idx_expenses_date ON expenses (expense_date) WHERE deleted_at IS NULL;

CREATE TABLE sessions (
    token_hash      TEXT PRIMARY KEY,
    admin_id        INTEGER NOT NULL REFERENCES admins(id),
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    expires_at      TEXT NOT NULL,
    last_seen_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    user_agent      TEXT,
    ip_address      TEXT
);

CREATE INDEX idx_sessions_admin ON sessions (admin_id);
CREATE INDEX idx_sessions_expires ON sessions (expires_at);
