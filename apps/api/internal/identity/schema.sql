CREATE TABLE IF NOT EXISTS player_accounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username text NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS player_accounts_username_ci
    ON player_accounts (lower(username));
CREATE UNIQUE INDEX IF NOT EXISTS player_accounts_email_ci
    ON player_accounts (lower(email));

CREATE TABLE IF NOT EXISTS player_sessions (
    token_hash bytea PRIMARY KEY,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS player_sessions_expiry ON player_sessions (expires_at);

CREATE TABLE IF NOT EXISTS player_characters (
    account_id uuid PRIMARY KEY REFERENCES player_accounts(id) ON DELETE CASCADE,
    display_name text NOT NULL,
    portrait_id smallint NOT NULL CHECK (portrait_id BETWEEN 1 AND 6),
    created_at timestamptz NOT NULL DEFAULT now()
);
