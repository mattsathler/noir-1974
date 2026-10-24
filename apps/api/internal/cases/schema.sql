CREATE TABLE IF NOT EXISTS cases (
    server_id uuid NOT NULL REFERENCES worlds(server_id) ON DELETE CASCADE,
    case_id text NOT NULL CHECK (case_id <> ''),
    case_seed text NOT NULL CHECK (case_seed <> ''),
    city_id text NOT NULL DEFAULT 'fr/lyon' CHECK (city_id <> ''),
    generator_version integer NOT NULL CHECK (generator_version > 0),
    case_type text NOT NULL CHECK (case_type = 'homicide'),
    archetype_id text NOT NULL DEFAULT 'workplace_conflict' CHECK (archetype_id <> ''),
    motive text NOT NULL DEFAULT 'Legacy workplace homicide.' CHECK (motive <> ''),
    victim_entity_id text NOT NULL,
    killer_entity_id text NOT NULL,
    location_entity_id text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, case_id),
    FOREIGN KEY (server_id, victim_entity_id)
        REFERENCES world_entities(server_id, entity_id),
    FOREIGN KEY (server_id, killer_entity_id)
        REFERENCES world_entities(server_id, entity_id),
    FOREIGN KEY (server_id, location_entity_id)
        REFERENCES world_entities(server_id, entity_id),
    CHECK (victim_entity_id <> killer_entity_id)
);
CREATE INDEX IF NOT EXISTS cases_by_server_created_idx
    ON cases(server_id, created_at);

ALTER TABLE cases ADD COLUMN IF NOT EXISTS archetype_id text NOT NULL DEFAULT 'workplace_conflict';
ALTER TABLE cases ADD COLUMN IF NOT EXISTS motive text NOT NULL DEFAULT 'Legacy workplace homicide.';
ALTER TABLE cases ADD COLUMN IF NOT EXISTS city_id text NOT NULL DEFAULT 'fr/lyon';
UPDATE cases c
SET city_id = substring(r.region_id from '^([^/]+/[^/]+)')
FROM world_entities e
JOIN world_regions r ON r.server_id = e.server_id AND r.region_id = e.region_id
WHERE c.server_id = e.server_id
  AND c.location_entity_id = e.entity_id
  AND c.city_id = 'fr/lyon'
  AND r.region_id LIKE '%/district/%';
CREATE INDEX IF NOT EXISTS cases_by_server_city_created_idx
    ON cases(server_id, city_id, created_at DESC);

CREATE TABLE IF NOT EXISTS case_relationships (
    server_id uuid NOT NULL,
    case_id text NOT NULL,
    relationship_id text NOT NULL,
    relationship_type text NOT NULL DEFAULT 'employment' CHECK (relationship_type IN ('employment', 'personal')),
    relationship_role text NOT NULL CHECK (relationship_role <> ''),
    PRIMARY KEY (server_id, case_id, relationship_id),
    UNIQUE (server_id, case_id, relationship_role),
    FOREIGN KEY (server_id, case_id)
        REFERENCES cases(server_id, case_id) ON DELETE CASCADE,
    CHECK (relationship_id <> '')
);

ALTER TABLE case_relationships ADD COLUMN IF NOT EXISTS relationship_type text NOT NULL DEFAULT 'employment';
ALTER TABLE case_relationships DROP CONSTRAINT IF EXISTS case_relationships_server_id_relationship_id_fkey;
ALTER TABLE case_relationships DROP CONSTRAINT IF EXISTS case_relationships_relationship_type_check;
ALTER TABLE case_relationships DROP CONSTRAINT IF EXISTS case_relationships_relationship_role_check;
ALTER TABLE case_relationships ADD CONSTRAINT case_relationships_relationship_type_check CHECK (relationship_type IN ('employment', 'personal'));
ALTER TABLE case_relationships ADD CONSTRAINT case_relationships_relationship_role_check CHECK (relationship_role <> '');

CREATE TABLE IF NOT EXISTS case_evidence (
    server_id uuid NOT NULL,
    case_id text NOT NULL,
    evidence_id text NOT NULL CHECK (evidence_id <> ''),
    location_entity_id text NOT NULL,
    evidence_type text NOT NULL CHECK (evidence_type <> ''),
    evidence_data jsonb NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY (server_id, case_id, evidence_id),
    FOREIGN KEY (server_id, case_id)
        REFERENCES cases(server_id, case_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, location_entity_id)
        REFERENCES world_entities(server_id, entity_id)
);
ALTER TABLE case_evidence DROP CONSTRAINT IF EXISTS case_evidence_evidence_type_check;
ALTER TABLE case_evidence ADD CONSTRAINT case_evidence_evidence_type_check CHECK (evidence_type <> '');
CREATE INDEX IF NOT EXISTS case_evidence_by_location_idx
    ON case_evidence(server_id, location_entity_id);

CREATE TABLE IF NOT EXISTS case_activity_records (
    server_id uuid NOT NULL,
    case_id text NOT NULL,
    record_id text NOT NULL CHECK (record_id <> ''),
    npc_entity_id text NOT NULL,
    related_npc_entity_id text,
    source text NOT NULL CHECK (source IN ('bank', 'gate')),
    occurred_at timestamptz NOT NULL,
    internal_relation text NOT NULL CHECK (internal_relation <> ''),
    record_data jsonb NOT NULL,
    PRIMARY KEY (server_id, record_id),
    FOREIGN KEY (server_id, case_id)
        REFERENCES cases(server_id, case_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, npc_entity_id)
        REFERENCES world_entities(server_id, entity_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, related_npc_entity_id)
        REFERENCES world_entities(server_id, entity_id) ON DELETE CASCADE
);
ALTER TABLE case_activity_records ADD COLUMN IF NOT EXISTS related_npc_entity_id text;
ALTER TABLE case_activity_records DROP CONSTRAINT IF EXISTS case_activity_records_server_id_related_npc_entity_id_fkey;
ALTER TABLE case_activity_records ADD CONSTRAINT case_activity_records_server_id_related_npc_entity_id_fkey
    FOREIGN KEY (server_id, related_npc_entity_id)
    REFERENCES world_entities(server_id, entity_id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS case_activity_records_by_npc_date_idx
    ON case_activity_records(server_id, npc_entity_id, occurred_at, record_id);
CREATE INDEX IF NOT EXISTS case_activity_records_by_npc_source_date_idx
    ON case_activity_records(server_id, npc_entity_id, source, occurred_at, record_id);

CREATE TABLE IF NOT EXISTS case_mail_messages (
    server_id uuid NOT NULL,
    case_id text NOT NULL,
    message_id text NOT NULL CHECK (message_id <> ''),
    mailbox_npc_entity_id text NOT NULL,
    related_npc_entity_id text NOT NULL,
    sent_at timestamptz NOT NULL,
    internal_relation text NOT NULL CHECK (internal_relation <> ''),
    message_data jsonb NOT NULL,
    PRIMARY KEY (server_id, message_id),
    FOREIGN KEY (server_id, case_id)
        REFERENCES cases(server_id, case_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, mailbox_npc_entity_id)
        REFERENCES world_entities(server_id, entity_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, related_npc_entity_id)
        REFERENCES world_entities(server_id, entity_id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS case_mail_messages_by_mailbox_date_idx
    ON case_mail_messages(server_id, mailbox_npc_entity_id, sent_at, message_id);
CREATE INDEX IF NOT EXISTS case_mail_messages_by_related_date_idx
    ON case_mail_messages(server_id, related_npc_entity_id, sent_at, message_id);

CREATE TABLE IF NOT EXISTS case_states (
    server_id uuid NOT NULL,
    case_id text NOT NULL,
    status text NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'solved', 'archived')),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, case_id),
    FOREIGN KEY (server_id, case_id)
        REFERENCES cases(server_id, case_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS player_case_progress (
    server_id uuid NOT NULL,
    case_id text NOT NULL,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    progress jsonb NOT NULL DEFAULT '{"discoveredEvidenceIds":[]}'::jsonb,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, case_id, account_id),
    FOREIGN KEY (server_id, case_id)
        REFERENCES cases(server_id, case_id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS player_case_progress_by_account_idx
    ON player_case_progress(account_id, server_id, updated_at);

CREATE TABLE IF NOT EXISTS player_search_warrants (
    server_id uuid NOT NULL REFERENCES worlds(server_id) ON DELETE CASCADE,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    warrant_id text NOT NULL CHECK (warrant_id <> ''),
    city_id text NOT NULL CHECK (city_id <> ''),
    station_location_id text NOT NULL CHECK (station_location_id <> ''),
    target_npc_id text NOT NULL CHECK (target_npc_id <> ''),
    status text NOT NULL CHECK (status IN ('pending', 'granted', 'denied')),
    catalog_version integer NOT NULL DEFAULT 2 CHECK (catalog_version > 0),
    ready_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, account_id, warrant_id)
);
ALTER TABLE player_search_warrants ADD COLUMN IF NOT EXISTS catalog_version integer NOT NULL DEFAULT 1;
ALTER TABLE player_search_warrants ADD COLUMN IF NOT EXISTS ready_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE player_search_warrants DROP CONSTRAINT IF EXISTS player_search_warrants_status_check;
ALTER TABLE player_search_warrants ADD CONSTRAINT player_search_warrants_status_check CHECK (status IN ('pending', 'granted', 'denied'));

CREATE TABLE IF NOT EXISTS player_async_requests (
    server_id uuid NOT NULL REFERENCES worlds(server_id) ON DELETE CASCADE,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    request_id text NOT NULL CHECK (request_id <> ''),
    request_type text NOT NULL CHECK (request_type IN ('residential_search', 'accusation')),
    request_key text NOT NULL CHECK (request_key <> ''),
    payload jsonb NOT NULL,
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed')),
    available_at timestamptz NOT NULL,
    lease_until timestamptz,
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz,
    PRIMARY KEY (server_id, request_id),
    UNIQUE (server_id, account_id, request_type, request_key)
);
CREATE INDEX IF NOT EXISTS player_async_requests_due_idx
    ON player_async_requests(server_id, available_at, created_at)
    WHERE status IN ('pending', 'processing');

CREATE TABLE IF NOT EXISTS player_inbox_messages (
    server_id uuid NOT NULL REFERENCES worlds(server_id) ON DELETE CASCADE,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    message_id text NOT NULL CHECK (message_id <> ''),
    warrant_id text,
    sent_at timestamptz NOT NULL DEFAULT now(),
    subject text NOT NULL CHECK (subject <> ''),
    body text NOT NULL,
    attachments jsonb NOT NULL DEFAULT '[]'::jsonb,
    PRIMARY KEY (server_id, account_id, message_id),
    FOREIGN KEY (server_id, account_id, warrant_id)
        REFERENCES player_search_warrants(server_id, account_id, warrant_id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS player_inbox_messages_by_date_idx
    ON player_inbox_messages(server_id, account_id, sent_at DESC);

CREATE TABLE IF NOT EXISTS player_case_accusations (
    server_id uuid NOT NULL,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    case_id text NOT NULL,
    board_id text NOT NULL,
    accepted boolean NOT NULL,
    outcome text NOT NULL CHECK (outcome IN ('arrested','insufficient')),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id,account_id,case_id),
    FOREIGN KEY (server_id,case_id) REFERENCES cases(server_id,case_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS player_local_access_attempts (
    server_id uuid NOT NULL REFERENCES worlds(server_id) ON DELETE CASCADE,
    attempt_id bigint GENERATED ALWAYS AS IDENTITY,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    city_id text NOT NULL CHECK (city_id <> ''),
    location_id text NOT NULL CHECK (location_id <> ''),
    company_id text NOT NULL DEFAULT '',
    facility_kind text NOT NULL CHECK (facility_kind <> ''),
    status text NOT NULL CHECK (status IN ('granted', 'denied')),
    attempted_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, attempt_id)
);
CREATE INDEX IF NOT EXISTS player_local_access_attempts_by_player_idx
    ON player_local_access_attempts(server_id, account_id, attempted_at DESC);

CREATE TABLE IF NOT EXISTS player_discovered_items (
    server_id uuid NOT NULL REFERENCES worlds(server_id) ON DELETE CASCADE,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    item_id text NOT NULL CHECK (item_id <> ''),
    item_source text NOT NULL CHECK (item_source IN ('email', 'bank', 'gate', 'case', 'person', 'company', 'note')),
    item_data jsonb NOT NULL,
    discovered_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, account_id, item_id)
);
ALTER TABLE player_discovered_items DROP CONSTRAINT IF EXISTS player_discovered_items_item_source_check;
ALTER TABLE player_discovered_items ADD CONSTRAINT player_discovered_items_item_source_check
    CHECK (item_source IN ('email', 'bank', 'gate', 'case', 'person', 'company', 'note'));

CREATE TABLE IF NOT EXISTS player_whiteboards (
    server_id uuid NOT NULL REFERENCES worlds(server_id) ON DELETE CASCADE,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    whiteboard_id text NOT NULL CHECK (whiteboard_id <> ''),
    title text NOT NULL CHECK (title <> ''),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, account_id, whiteboard_id)
);
CREATE INDEX IF NOT EXISTS player_whiteboards_by_owner_idx
    ON player_whiteboards(server_id, account_id, created_at, whiteboard_id);

CREATE TABLE IF NOT EXISTS player_whiteboard_pins (
    server_id uuid NOT NULL,
    account_id uuid NOT NULL,
    whiteboard_id text NOT NULL,
    item_id text NOT NULL,
    designation text NOT NULL DEFAULT '' CHECK (designation IN ('', 'proof', 'accused', 'victim')),
    position_x integer NOT NULL DEFAULT 80 CHECK (position_x >= 0),
    position_y integer NOT NULL DEFAULT 80 CHECK (position_y >= 0),
    pinned_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, account_id, whiteboard_id, item_id),
    FOREIGN KEY (server_id, account_id, whiteboard_id)
        REFERENCES player_whiteboards(server_id, account_id, whiteboard_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, account_id, item_id)
        REFERENCES player_discovered_items(server_id, account_id, item_id) ON DELETE CASCADE
);
ALTER TABLE player_whiteboard_pins ADD COLUMN IF NOT EXISTS designation text NOT NULL DEFAULT '';
ALTER TABLE player_whiteboard_pins DROP CONSTRAINT IF EXISTS player_whiteboard_pins_designation_check;
ALTER TABLE player_whiteboard_pins ADD CONSTRAINT player_whiteboard_pins_designation_check CHECK (designation IN ('', 'proof', 'accused', 'victim'));
CREATE UNIQUE INDEX IF NOT EXISTS player_whiteboard_one_accused_idx ON player_whiteboard_pins(server_id,account_id,whiteboard_id) WHERE designation='accused';
CREATE UNIQUE INDEX IF NOT EXISTS player_whiteboard_one_victim_idx ON player_whiteboard_pins(server_id,account_id,whiteboard_id) WHERE designation='victim';

CREATE TABLE IF NOT EXISTS player_whiteboard_connections (
    server_id uuid NOT NULL,
    account_id uuid NOT NULL,
    whiteboard_id text NOT NULL,
    from_item_id text NOT NULL,
    to_item_id text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, account_id, whiteboard_id, from_item_id, to_item_id),
    FOREIGN KEY (server_id, account_id, whiteboard_id, from_item_id)
        REFERENCES player_whiteboard_pins(server_id, account_id, whiteboard_id, item_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, account_id, whiteboard_id, to_item_id)
        REFERENCES player_whiteboard_pins(server_id, account_id, whiteboard_id, item_id) ON DELETE CASCADE,
    CHECK (from_item_id < to_item_id)
);
CREATE TABLE IF NOT EXISTS case_generation_schedule (
    server_id uuid PRIMARY KEY REFERENCES worlds(server_id) ON DELETE CASCADE,
    next_slot bigint NOT NULL DEFAULT 1,
    next_run_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);
