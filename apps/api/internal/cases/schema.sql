CREATE TABLE IF NOT EXISTS cases (
    server_id uuid NOT NULL REFERENCES worlds(server_id) ON DELETE CASCADE,
    case_id text NOT NULL CHECK (case_id <> ''),
    case_seed text NOT NULL CHECK (case_seed <> ''),
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
