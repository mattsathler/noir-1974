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
    visual_identity jsonb NOT NULL DEFAULT '{"hairId":"hair_short_brown","skinId":"skin_white","eyesId":"eyes_brown","faceId":"","beardId":"","accessoryIds":[]}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS player_economy (
    server_id uuid NOT NULL,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    fame bigint NOT NULL DEFAULT 0 CHECK (fame >= 0),
    credits_cents bigint NOT NULL DEFAULT 100000 CHECK (credits_cents >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, account_id)
);

CREATE TABLE IF NOT EXISTS player_stock_positions (
    server_id uuid NOT NULL,
    account_id uuid NOT NULL,
    company_id text NOT NULL CHECK (company_id <> ''),
    country_id text NOT NULL CHECK (country_id <> ''),
    company_name text NOT NULL CHECK (company_name <> ''),
    city_name text NOT NULL DEFAULT '',
    shares bigint NOT NULL CHECK (shares > 0),
    average_cost_cents bigint NOT NULL CHECK (average_cost_cents > 0),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, account_id, company_id),
    FOREIGN KEY (server_id, account_id)
        REFERENCES player_economy(server_id, account_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS player_stock_orders (
    order_id bigserial PRIMARY KEY,
    server_id uuid NOT NULL,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    client_order_id text NOT NULL CHECK (client_order_id <> ''),
    company_id text NOT NULL CHECK (company_id <> ''),
    side text NOT NULL CHECK (side IN ('buy', 'sell')),
    shares bigint NOT NULL CHECK (shares > 0),
    price_cents bigint NOT NULL CHECK (price_cents > 0),
    total_cents bigint NOT NULL CHECK (total_cents > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (server_id, account_id, client_order_id)
);

CREATE TABLE IF NOT EXISTS company_market_scores (
    server_id uuid NOT NULL,
    company_id text NOT NULL CHECK (company_id <> ''),
    score integer NOT NULL DEFAULT 1000 CHECK (score BETWEEN 500 AND 1500),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, company_id)
);

CREATE TABLE IF NOT EXISTS company_market_events (
    server_id uuid NOT NULL,
    event_id text NOT NULL CHECK (event_id <> ''),
    company_id text NOT NULL CHECK (company_id <> ''),
    event_type text NOT NULL CHECK (event_type IN ('workplace_conflict', 'financial_extortion', 'blackwire_leak')),
    score_delta integer NOT NULL,
    score_before integer NOT NULL CHECK (score_before BETWEEN 500 AND 1500),
    score_after integer NOT NULL CHECK (score_after BETWEEN 500 AND 1500),
    occurred_at timestamptz NOT NULL,
    recorded_at timestamptz NOT NULL DEFAULT now(),
    correlation_id text NOT NULL DEFAULT '',
    source_company_id text NOT NULL DEFAULT '',
    PRIMARY KEY (server_id, event_id),
    FOREIGN KEY (server_id, company_id)
        REFERENCES company_market_scores(server_id, company_id) ON DELETE CASCADE
);
ALTER TABLE company_market_events DROP CONSTRAINT IF EXISTS company_market_events_event_type_check;
ALTER TABLE company_market_events ADD COLUMN IF NOT EXISTS correlation_id text NOT NULL DEFAULT '';
ALTER TABLE company_market_events ADD COLUMN IF NOT EXISTS source_company_id text NOT NULL DEFAULT '';
ALTER TABLE company_market_events ADD CONSTRAINT company_market_events_event_type_check
    CHECK (event_type IN ('workplace_conflict', 'financial_extortion', 'blackwire_leak', 'rival_market_movement'));
CREATE INDEX IF NOT EXISTS company_market_events_history_idx
    ON company_market_events(server_id, company_id, recorded_at DESC, event_id);
CREATE INDEX IF NOT EXISTS company_market_events_correlation_idx
    ON company_market_events(server_id, correlation_id, company_id);

CREATE TABLE IF NOT EXISTS player_blackwire_sales (
    server_id uuid NOT NULL,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    item_id text NOT NULL CHECK (item_id <> ''),
    whiteboard_id text NOT NULL CHECK (whiteboard_id <> ''),
    price_cents bigint NOT NULL CHECK (price_cents > 0),
    category text NOT NULL CHECK (category <> ''),
    company_id text,
    sold_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, account_id, item_id),
    FOREIGN KEY (server_id, account_id) REFERENCES player_economy(server_id, account_id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS player_blackwire_sales_history_idx
    ON player_blackwire_sales(server_id, account_id, sold_at DESC);
ALTER TABLE player_characters ADD COLUMN IF NOT EXISTS visual_identity jsonb NOT NULL DEFAULT '{"hairId":"hair_short_brown","skinId":"skin_white","eyesId":"eyes_brown","faceId":"","beardId":"","accessoryIds":[]}'::jsonb;
ALTER TABLE player_characters ALTER COLUMN visual_identity SET DEFAULT '{"hairId":"hair_short_brown","skinId":"skin_white","eyesId":"eyes_brown","faceId":"","beardId":"","accessoryIds":[]}'::jsonb;

UPDATE player_characters
SET visual_identity = (visual_identity - 'faceFeatureIds') || jsonb_build_object('faceDetailId', COALESCE(visual_identity->'faceFeatureIds'->>0, ''))
WHERE visual_identity ? 'faceFeatureIds' AND NOT (visual_identity ? 'faceDetailId');

UPDATE player_characters
SET visual_identity = jsonb_set(
    jsonb_set(
        visual_identity,
        '{hairId}',
        to_jsonb(regexp_replace(visual_identity->>'hairId', '^hair_(bob|cropped)_', 'hair_short_')),
        true
    ),
    '{faceDetailId}',
    to_jsonb(CASE
        WHEN visual_identity->>'faceDetailId' = 'face_beard_01' THEN
            'face_beard_' || COALESCE(NULLIF(split_part(regexp_replace(visual_identity->>'hairId', '^hair_(bob|cropped)_', ''), '_', 1), ''), 'brown') || '_01'
        WHEN visual_identity->>'faceDetailId' = 'face_moustache_01' THEN 'face_moustache_black_01'
        ELSE COALESCE(visual_identity->>'faceDetailId', '')
    END),
    true
)
WHERE visual_identity->>'hairId' ~ '^hair_(bob|cropped)_'
   OR visual_identity->>'faceDetailId' IN ('face_beard_01', 'face_moustache_01');

UPDATE player_characters
SET visual_identity = jsonb_set(
    visual_identity,
    '{accessoryIds}',
    (
        SELECT COALESCE(jsonb_agg(to_jsonb(CASE value
            WHEN 'face_glasses_01' THEN 'face_glasses_black_01'
            WHEN 'face_glasses_02' THEN 'face_glasses_black_02'
            ELSE value
        END)), '[]'::jsonb)
        FROM jsonb_array_elements_text(CASE
            WHEN jsonb_typeof(visual_identity->'accessoryIds') = 'array' THEN visual_identity->'accessoryIds'
            ELSE '[]'::jsonb
        END) AS item(value)
    ),
    true
)
WHERE visual_identity->'accessoryIds' @> '["face_glasses_01"]'::jsonb
   OR visual_identity->'accessoryIds' @> '["face_glasses_02"]'::jsonb;

-- Remove numeric suffixes from persisted visual-layer IDs and name multi-style items.
UPDATE player_characters
SET visual_identity = jsonb_set(
    jsonb_set(
        jsonb_set(
            jsonb_set(
                visual_identity,
                '{hairId}',
                to_jsonb(CASE
                    WHEN visual_identity->>'hairId' ~ '^hair_short_(black|brown|ginger|blonde|grey)_02$'
                        THEN regexp_replace(visual_identity->>'hairId', '^hair_short_(.*)_02$', 'hair_short_alt_\1')
                    WHEN visual_identity->>'hairId' = 'hair_short_01' THEN 'hair_short_base'
                    ELSE regexp_replace(COALESCE(visual_identity->>'hairId', ''), '_01$', '')
                END),
                true
            ),
            '{skinColorId}',
            to_jsonb(regexp_replace(COALESCE(visual_identity->>'skinColorId', ''), '_01$', '')),
            true
        ),
        '{eyeColorId}',
        to_jsonb(regexp_replace(COALESCE(visual_identity->>'eyeColorId', ''), '_01$', '')),
        true
    ),
    '{faceDetailId}',
    to_jsonb(regexp_replace(COALESCE(visual_identity->>'faceDetailId', ''), '_01$', '')),
    true
)
WHERE visual_identity IS NOT NULL;

UPDATE player_characters
SET visual_identity = jsonb_set(
    visual_identity,
    '{accessoryIds}',
    (
        SELECT COALESCE(jsonb_agg(to_jsonb(CASE
            WHEN value ~ '^face_glasses_(black|tortoise|brass|silver|ivory)_01$'
                THEN regexp_replace(value, '^face_glasses_(.*)_01$', 'face_glasses_round_\1')
            WHEN value ~ '^face_glasses_(black|tortoise|brass|silver|ivory)_02$'
                THEN regexp_replace(value, '^face_glasses_(.*)_02$', 'face_glasses_square_\1')
            ELSE regexp_replace(value, '_01$', '')
        END)), '[]'::jsonb)
        FROM jsonb_array_elements_text(CASE
            WHEN jsonb_typeof(visual_identity->'accessoryIds') = 'array' THEN visual_identity->'accessoryIds'
            ELSE '[]'::jsonb
        END) AS item(value)
    ),
    true
)
WHERE visual_identity ? 'accessoryIds';

CREATE TABLE IF NOT EXISTS player_locations (
    account_id uuid PRIMARY KEY REFERENCES player_accounts(id) ON DELETE CASCADE,
    location_id text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);
