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
    visual_identity jsonb NOT NULL DEFAULT '{"hairId":"hair_short_brown","skinColorId":"skin_light","eyeColorId":"eyes_brown","faceDetailId":"","accessoryIds":[]}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE player_characters ADD COLUMN IF NOT EXISTS visual_identity jsonb NOT NULL DEFAULT '{"hairId":"hair_short_brown","skinColorId":"skin_light","eyeColorId":"eyes_brown","faceDetailId":"","accessoryIds":[]}'::jsonb;
ALTER TABLE player_characters ALTER COLUMN visual_identity SET DEFAULT '{"hairId":"hair_short_brown","skinColorId":"skin_light","eyeColorId":"eyes_brown","faceDetailId":"","accessoryIds":[]}'::jsonb;

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
        FROM jsonb_array_elements_text(visual_identity->'accessoryIds') AS item(value)
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
        FROM jsonb_array_elements_text(COALESCE(visual_identity->'accessoryIds', '[]'::jsonb)) AS item(value)
    ),
    true
)
WHERE visual_identity ? 'accessoryIds';

CREATE TABLE IF NOT EXISTS player_locations (
    account_id uuid PRIMARY KEY REFERENCES player_accounts(id) ON DELETE CASCADE,
    location_id text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);
