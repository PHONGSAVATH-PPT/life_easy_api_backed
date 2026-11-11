CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(254) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE families (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE family_members (
    family_id UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    role TEXT NOT NULL CHECK (role IN ('owner', 'member')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (family_id, user_id),
    UNIQUE (user_id)
);

CREATE TABLE entries (
    id TEXT PRIMARY KEY CHECK (LENGTH(id) BETWEEN 1 AND 100),
    owner_id UUID NOT NULL REFERENCES users(id),
    family_id UUID REFERENCES families(id),
    kind TEXT NOT NULL CHECK (
        kind IN ('task', 'appointment', 'income', 'expense', 'bill', 'document')
    ),
    title VARCHAR(200) NOT NULL CHECK (LENGTH(TRIM(title)) > 0),
    due_at TIMESTAMPTZ NOT NULL,
    amount_minor BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'LAK'
        CHECK (currency IN ('LAK', 'THB')),
    done BOOLEAN NOT NULL DEFAULT FALSE,
    shared BOOLEAN NOT NULL DEFAULT FALSE,
    note TEXT NOT NULL DEFAULT '',
    assignee_id UUID REFERENCES users(id),
    recurrence TEXT NOT NULL DEFAULT 'none'
        CHECK (recurrence IN ('none', 'monthly')),
    reminder_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    source_bill_id TEXT UNIQUE REFERENCES entries(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (shared = (family_id IS NOT NULL)),

    CHECK (
        (kind IN ('income', 'expense', 'bill')
            AND amount_minor > 0
            AND amount_minor <= 90000000000000)
        OR
        (kind NOT IN ('income', 'expense', 'bill')
            AND amount_minor = 0)
    ),

    CHECK (source_bill_id IS NULL OR kind = 'expense'),

    FOREIGN KEY (family_id, assignee_id)
        REFERENCES family_members(family_id, user_id)
        ON DELETE SET NULL (assignee_id)
);

CREATE INDEX entries_owner_due_idx
    ON entries(owner_id, due_at DESC);

CREATE INDEX entries_shared_family_due_idx
    ON entries(family_id, due_at)
    WHERE shared = TRUE;

CREATE INDEX entries_owner_kind_due_idx
    ON entries(owner_id, kind, due_at DESC);