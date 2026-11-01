CREATE TABLE IF NOT EXISTS cases (
    server_id uuid NOT NULL REFERENCES worlds(server_id) ON DELETE CASCADE,
    case_id text NOT NULL CHECK (case_id <> ''),
    case_seed text NOT NULL CHECK (case_seed <> ''),
    city_id text NOT NULL DEFAULT 'fr/lyon' CHECK (city_id <> ''),
    generator_version integer NOT NULL CHECK (generator_version > 0),
    case_type text NOT NULL CHECK (case_type IN ('homicide', 'embezzlement', 'financial_extortion', 'accounting_fraud')),
    archetype_id text NOT NULL DEFAULT 'workplace_conflict' CHECK (archetype_id <> ''),
    company_id text,
    motive text NOT NULL DEFAULT 'Legacy workplace homicide.' CHECK (motive <> ''),
    victim_entity_id text,
    killer_entity_id text NOT NULL,
    location_entity_id text,
    crime_at timestamptz,
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
ALTER TABLE cases ADD COLUMN IF NOT EXISTS company_id text;
ALTER TABLE cases ADD COLUMN IF NOT EXISTS motive text NOT NULL DEFAULT 'Legacy workplace homicide.';
ALTER TABLE cases ADD COLUMN IF NOT EXISTS city_id text NOT NULL DEFAULT 'fr/lyon';
ALTER TABLE cases ADD COLUMN IF NOT EXISTS crime_at timestamptz;
ALTER TABLE cases ALTER COLUMN victim_entity_id DROP NOT NULL;
ALTER TABLE cases ALTER COLUMN location_entity_id DROP NOT NULL;
ALTER TABLE cases DROP CONSTRAINT IF EXISTS cases_case_type_check;
ALTER TABLE cases ADD CONSTRAINT cases_case_type_check CHECK (case_type IN ('homicide', 'embezzlement', 'financial_extortion', 'accounting_fraud'));
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

CREATE TABLE IF NOT EXISTS case_participants (
    server_id uuid NOT NULL,
    case_id text NOT NULL,
    npc_entity_id text NOT NULL,
    participant_role text NOT NULL CHECK (participant_role IN ('principal', 'accomplice', 'victim')),
    PRIMARY KEY (server_id, case_id, npc_entity_id),
    FOREIGN KEY (server_id, case_id) REFERENCES cases(server_id, case_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, npc_entity_id) REFERENCES world_entities(server_id, entity_id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS case_participants_by_npc_idx ON case_participants(server_id, npc_entity_id, case_id);

INSERT INTO case_participants(server_id,case_id,npc_entity_id,participant_role)
SELECT server_id,case_id,killer_entity_id,'principal' FROM cases
ON CONFLICT DO NOTHING;
INSERT INTO case_participants(server_id,case_id,npc_entity_id,participant_role)
SELECT server_id,case_id,victim_entity_id,'victim' FROM cases WHERE victim_entity_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- A case is an investigation; each criminal occurrence is stored independently.
CREATE TABLE IF NOT EXISTS case_crimes (
    server_id uuid NOT NULL,
    case_id text NOT NULL,
    crime_id text NOT NULL CHECK (crime_id <> ''),
    crime_type text NOT NULL CHECK (crime_type <> ''),
    occurred_at timestamptz NOT NULL,
    location_entity_id text,
    parent_crime_id text,
    crime_sequence integer NOT NULL CHECK (crime_sequence > 0),
    source_key text NOT NULL CHECK (source_key <> ''),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, crime_id),
    UNIQUE (server_id, case_id, crime_id),
    UNIQUE (server_id, case_id, crime_sequence),
    UNIQUE (server_id, case_id, source_key),
    FOREIGN KEY (server_id, case_id, parent_crime_id) REFERENCES case_crimes(server_id, case_id, crime_id),
    FOREIGN KEY (server_id, case_id) REFERENCES cases(server_id, case_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, location_entity_id) REFERENCES world_entities(server_id, entity_id)
);
ALTER TABLE case_crimes ADD COLUMN IF NOT EXISTS parent_crime_id text;
ALTER TABLE case_crimes DROP CONSTRAINT IF EXISTS case_crimes_server_id_case_id_parent_crime_id_fkey;
ALTER TABLE case_crimes ADD CONSTRAINT case_crimes_server_id_case_id_parent_crime_id_fkey
    FOREIGN KEY (server_id, case_id, parent_crime_id) REFERENCES case_crimes(server_id, case_id, crime_id);
CREATE INDEX IF NOT EXISTS case_crimes_by_case_time_idx
    ON case_crimes(server_id, case_id, occurred_at, crime_sequence);

CREATE TABLE IF NOT EXISTS case_crime_participants (
    server_id uuid NOT NULL,
    crime_id text NOT NULL,
    npc_entity_id text NOT NULL,
    participant_role text NOT NULL CHECK (participant_role IN ('principal', 'accomplice', 'victim')),
    PRIMARY KEY (server_id, crime_id, npc_entity_id),
    FOREIGN KEY (server_id, crime_id) REFERENCES case_crimes(server_id, crime_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, npc_entity_id) REFERENCES world_entities(server_id, entity_id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS case_crime_participants_by_npc_idx
    ON case_crime_participants(server_id, npc_entity_id, crime_id);

-- Upgrade existing cases into a first persisted occurrence without changing
-- their public representation or duplicating the event on later startups.
INSERT INTO case_crimes(server_id, case_id, crime_id, crime_type, occurred_at, location_entity_id, crime_sequence, source_key)
SELECT server_id, case_id, case_id || '/crime/1', case_type, COALESCE(crime_at, created_at), location_entity_id, 1, 'primary'
FROM cases
ON CONFLICT DO NOTHING;
INSERT INTO case_crime_participants(server_id, crime_id, npc_entity_id, participant_role)
SELECT p.server_id, c.crime_id, p.npc_entity_id, p.participant_role
FROM case_participants p
JOIN case_crimes c ON c.server_id=p.server_id AND c.case_id=p.case_id AND c.source_key='primary'
ON CONFLICT DO NOTHING;

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
    crime_id text NOT NULL,
    evidence_id text NOT NULL CHECK (evidence_id <> ''),
    location_entity_id text,
    evidence_type text NOT NULL CHECK (evidence_type <> ''),
    evidence_data jsonb NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY (server_id, case_id, evidence_id),
    FOREIGN KEY (server_id, case_id)
        REFERENCES cases(server_id, case_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, case_id, crime_id)
        REFERENCES case_crimes(server_id, case_id, crime_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, location_entity_id)
        REFERENCES world_entities(server_id, entity_id)
);
ALTER TABLE case_evidence ADD COLUMN IF NOT EXISTS crime_id text;
UPDATE case_evidence e SET crime_id=c.crime_id
FROM case_crimes c
WHERE e.server_id=c.server_id AND e.case_id=c.case_id AND c.source_key='primary' AND e.crime_id IS NULL;
ALTER TABLE case_evidence ALTER COLUMN crime_id SET NOT NULL;
ALTER TABLE case_evidence DROP CONSTRAINT IF EXISTS case_evidence_server_id_case_id_crime_id_fkey;
ALTER TABLE case_evidence ADD CONSTRAINT case_evidence_server_id_case_id_crime_id_fkey
    FOREIGN KEY (server_id, case_id, crime_id)
    REFERENCES case_crimes(server_id, case_id, crime_id) ON DELETE CASCADE;
ALTER TABLE case_evidence DROP CONSTRAINT IF EXISTS case_evidence_evidence_type_check;
ALTER TABLE case_evidence ADD CONSTRAINT case_evidence_evidence_type_check CHECK (evidence_type <> '');
CREATE INDEX IF NOT EXISTS case_evidence_by_location_idx
    ON case_evidence(server_id, location_entity_id);
ALTER TABLE case_evidence ALTER COLUMN location_entity_id DROP NOT NULL;

CREATE TABLE IF NOT EXISTS case_activity_records (
    server_id uuid NOT NULL,
    case_id text NOT NULL,
    crime_id text NOT NULL,
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
    FOREIGN KEY (server_id, case_id, crime_id)
        REFERENCES case_crimes(server_id, case_id, crime_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, npc_entity_id)
        REFERENCES world_entities(server_id, entity_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, related_npc_entity_id)
        REFERENCES world_entities(server_id, entity_id) ON DELETE CASCADE
);
ALTER TABLE case_activity_records ADD COLUMN IF NOT EXISTS crime_id text;
UPDATE case_activity_records r SET crime_id=c.crime_id
FROM case_crimes c
WHERE r.server_id=c.server_id AND r.case_id=c.case_id AND c.source_key='primary' AND r.crime_id IS NULL;
ALTER TABLE case_activity_records ALTER COLUMN crime_id SET NOT NULL;
ALTER TABLE case_activity_records DROP CONSTRAINT IF EXISTS case_activity_records_server_id_case_id_crime_id_fkey;
ALTER TABLE case_activity_records ADD CONSTRAINT case_activity_records_server_id_case_id_crime_id_fkey
    FOREIGN KEY (server_id, case_id, crime_id)
    REFERENCES case_crimes(server_id, case_id, crime_id) ON DELETE CASCADE;
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
    crime_id text NOT NULL,
    message_id text NOT NULL CHECK (message_id <> ''),
    mailbox_npc_entity_id text NOT NULL,
    related_npc_entity_id text NOT NULL,
    sent_at timestamptz NOT NULL,
    internal_relation text NOT NULL CHECK (internal_relation <> ''),
    message_data jsonb NOT NULL,
    PRIMARY KEY (server_id, message_id),
    FOREIGN KEY (server_id, case_id)
        REFERENCES cases(server_id, case_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, case_id, crime_id)
        REFERENCES case_crimes(server_id, case_id, crime_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, mailbox_npc_entity_id)
        REFERENCES world_entities(server_id, entity_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, related_npc_entity_id)
        REFERENCES world_entities(server_id, entity_id) ON DELETE CASCADE
);
ALTER TABLE case_mail_messages ADD COLUMN IF NOT EXISTS crime_id text;
UPDATE case_mail_messages m SET crime_id=c.crime_id
FROM case_crimes c
WHERE m.server_id=c.server_id AND m.case_id=c.case_id AND c.source_key='primary' AND m.crime_id IS NULL;
ALTER TABLE case_mail_messages ALTER COLUMN crime_id SET NOT NULL;
ALTER TABLE case_mail_messages DROP CONSTRAINT IF EXISTS case_mail_messages_server_id_case_id_crime_id_fkey;
ALTER TABLE case_mail_messages ADD CONSTRAINT case_mail_messages_server_id_case_id_crime_id_fkey
    FOREIGN KEY (server_id, case_id, crime_id)
    REFERENCES case_crimes(server_id, case_id, crime_id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS case_mail_messages_by_mailbox_date_idx
    ON case_mail_messages(server_id, mailbox_npc_entity_id, sent_at, message_id);
CREATE INDEX IF NOT EXISTS case_mail_messages_by_related_date_idx
    ON case_mail_messages(server_id, related_npc_entity_id, sent_at, message_id);

CREATE TABLE IF NOT EXISTS case_states (
    server_id uuid NOT NULL,
    case_id text NOT NULL,
    status text NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'solved', 'archived')),
    updated_at timestamptz NOT NULL DEFAULT now(),
    body_found_at timestamptz,
    body_found_method text CHECK (body_found_method IN ('automatic', 'residential_search')),
    PRIMARY KEY (server_id, case_id),
    FOREIGN KEY (server_id, case_id)
        REFERENCES cases(server_id, case_id) ON DELETE CASCADE
);
ALTER TABLE case_states ADD COLUMN IF NOT EXISTS body_found_at timestamptz;
ALTER TABLE case_states ADD COLUMN IF NOT EXISTS body_found_method text;
ALTER TABLE case_states DROP CONSTRAINT IF EXISTS case_states_body_found_method_check;
ALTER TABLE case_states ADD CONSTRAINT case_states_body_found_method_check CHECK (body_found_method IS NULL OR body_found_method IN ('automatic', 'residential_search'));
CREATE INDEX IF NOT EXISTS case_states_body_discovery_idx ON case_states(server_id, body_found_at) WHERE body_found_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS case_evolution_jobs (
    server_id uuid NOT NULL,
    case_id text NOT NULL,
    event_key text NOT NULL CHECK (event_key <> ''),
    event_type text NOT NULL CHECK (event_type <> ''),
    plan_version integer NOT NULL DEFAULT 1 CHECK (plan_version > 0),
    available_at timestamptz NOT NULL,
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'cancelled')),
    completed_at timestamptz,
    PRIMARY KEY (server_id, case_id, event_key),
    FOREIGN KEY (server_id, case_id) REFERENCES cases(server_id, case_id) ON DELETE CASCADE
);
ALTER TABLE case_evolution_jobs ADD COLUMN IF NOT EXISTS plan_version integer NOT NULL DEFAULT 1;
CREATE INDEX IF NOT EXISTS case_evolution_jobs_due_idx
    ON case_evolution_jobs(server_id, available_at, case_id) WHERE status='pending';
INSERT INTO case_evolution_jobs(server_id,case_id,event_key,event_type,available_at)
SELECT c.server_id,c.case_id,'financial-cover-up','financial_cover_up',now()+interval '3 minutes'
FROM cases c JOIN case_states s USING(server_id,case_id)
WHERE c.case_type IN ('embezzlement','financial_extortion','accounting_fraud') AND s.status='open'
ON CONFLICT DO NOTHING;
UPDATE case_states s SET body_found_at=COALESCE(c.crime_at,c.created_at)+interval '90 minutes',body_found_method='automatic'
FROM cases c JOIN world_entities scene ON scene.server_id=c.server_id AND scene.entity_id=c.location_entity_id
WHERE s.server_id=c.server_id AND s.case_id=c.case_id AND s.body_found_at IS NULL
  AND c.case_type='homicide'
  AND scene.generation_key !~* '(^|/)(residence|residential|apartment|house|home|immeuble|flat)(/|[-_]|$)|/building/(residence|residential|apartment|house|home|immeuble|flat)([-_/]|$)';
INSERT INTO world_mutations(server_id,entity_id,field_key,value)
SELECT c.server_id,c.victim_entity_id,'status','"missing"'::jsonb
FROM cases c JOIN case_states s USING(server_id,case_id)
WHERE c.victim_entity_id IS NOT NULL AND s.status IN ('open','solved') AND COALESCE(c.crime_at,c.created_at)<=now() AND (s.body_found_at IS NULL OR s.body_found_at>now())
  AND (SELECT value #>> '{}' FROM world_mutations m WHERE m.server_id=c.server_id AND m.entity_id=c.victim_entity_id AND m.field_key='status' ORDER BY mutation_order DESC LIMIT 1)='dead';

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

CREATE TABLE IF NOT EXISTS player_email_breaker_access (
    server_id uuid NOT NULL REFERENCES worlds(server_id) ON DELETE CASCADE,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    request_id text NOT NULL CHECK (request_id <> ''),
    target_npc_id text NOT NULL CHECK (target_npc_id <> ''),
    target_name text NOT NULL CHECK (target_name <> ''),
    target_email text NOT NULL CHECK (target_email <> ''),
    from_date date NOT NULL,
    through_date date NOT NULL,
    ready_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, account_id, request_id),
    CHECK (through_date >= from_date)
);

CREATE TABLE IF NOT EXISTS player_bankrupt_access (
    server_id uuid NOT NULL REFERENCES worlds(server_id) ON DELETE CASCADE,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    target_npc_id text NOT NULL CHECK (target_npc_id <> ''),
    target_account_number text NOT NULL CHECK (target_account_number ~ '^[0-9]{8}$'),
    target_name text NOT NULL CHECK (target_name <> ''),
    opened_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, account_id, target_npc_id),
    UNIQUE (server_id, account_id, target_account_number)
);
CREATE INDEX IF NOT EXISTS player_bankrupt_access_by_player_idx
    ON player_bankrupt_access(server_id, account_id, opened_at DESC);
CREATE INDEX IF NOT EXISTS player_email_breaker_access_by_player_idx
    ON player_email_breaker_access(server_id, account_id, ready_at DESC);
CREATE INDEX IF NOT EXISTS player_async_requests_due_idx
    ON player_async_requests(server_id, available_at, created_at)
    WHERE status IN ('pending', 'processing');

CREATE TABLE IF NOT EXISTS player_inbox_messages (
    server_id uuid NOT NULL REFERENCES worlds(server_id) ON DELETE CASCADE,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    case_id text,
    crime_id text,
    message_id text NOT NULL CHECK (message_id <> ''),
    warrant_id text,
    sent_at timestamptz NOT NULL DEFAULT now(),
    sender text NOT NULL DEFAULT 'BILL BOUER',
    message_category text NOT NULL DEFAULT 'official' CHECK (message_category IN ('official','spam')),
    subject text NOT NULL CHECK (subject <> ''),
    body text NOT NULL,
    attachments jsonb NOT NULL DEFAULT '[]'::jsonb,
    deleted_at timestamptz,
    PRIMARY KEY (server_id, account_id, message_id),
    FOREIGN KEY (server_id, account_id, warrant_id)
        REFERENCES player_search_warrants(server_id, account_id, warrant_id) ON DELETE CASCADE,
    FOREIGN KEY (server_id, case_id) REFERENCES cases(server_id, case_id) ON DELETE CASCADE
);
ALTER TABLE player_inbox_messages ADD COLUMN IF NOT EXISTS case_id text;
ALTER TABLE player_inbox_messages ADD COLUMN IF NOT EXISTS crime_id text;
ALTER TABLE player_inbox_messages DROP CONSTRAINT IF EXISTS player_inbox_messages_server_id_case_id_fkey;
ALTER TABLE player_inbox_messages ADD CONSTRAINT player_inbox_messages_server_id_case_id_fkey
    FOREIGN KEY (server_id, case_id) REFERENCES cases(server_id, case_id) ON DELETE CASCADE;
ALTER TABLE player_inbox_messages DROP CONSTRAINT IF EXISTS player_inbox_messages_server_id_case_id_crime_id_fkey;
ALTER TABLE player_inbox_messages ADD CONSTRAINT player_inbox_messages_server_id_case_id_crime_id_fkey
    FOREIGN KEY (server_id, case_id, crime_id) REFERENCES case_crimes(server_id, case_id, crime_id) ON DELETE CASCADE;
ALTER TABLE player_inbox_messages ADD COLUMN IF NOT EXISTS sender text NOT NULL DEFAULT 'BILL BOUER';
ALTER TABLE player_inbox_messages ADD COLUMN IF NOT EXISTS message_category text NOT NULL DEFAULT 'official';
ALTER TABLE player_inbox_messages DROP CONSTRAINT IF EXISTS player_inbox_messages_message_category_check;
ALTER TABLE player_inbox_messages ADD CONSTRAINT player_inbox_messages_message_category_check CHECK (message_category IN ('official','spam'));
ALTER TABLE player_inbox_messages ADD COLUMN IF NOT EXISTS deleted_at timestamptz;
CREATE INDEX IF NOT EXISTS player_inbox_messages_by_date_idx
    ON player_inbox_messages(server_id, account_id, sent_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS player_npc_email_questions (
    server_id uuid NOT NULL REFERENCES worlds(server_id) ON DELETE CASCADE,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    request_id text NOT NULL CHECK (request_id <> ''),
    target_npc_id text NOT NULL CHECK (target_npc_id <> ''),
    target_name text NOT NULL CHECK (target_name <> ''),
    target_email text NOT NULL CHECK (target_email <> ''),
    question_id text NOT NULL CHECK (question_id IN ('routine','associates','unusual','work','worries')),
    subject text NOT NULL CHECK (subject <> ''),
    reply_body text NOT NULL CHECK (reply_body <> ''),
    reply_generated boolean NOT NULL DEFAULT true,
    ready_at timestamptz NOT NULL,
    delivered_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, account_id, request_id)
);
ALTER TABLE player_npc_email_questions ADD COLUMN IF NOT EXISTS reply_generated boolean NOT NULL DEFAULT true;
UPDATE player_npc_email_questions SET reply_generated=false WHERE reply_body='[no reply generated]' AND reply_generated=true;
CREATE INDEX IF NOT EXISTS player_npc_email_questions_due_idx
    ON player_npc_email_questions(server_id, account_id, ready_at)
    WHERE delivered_at IS NULL;
CREATE INDEX IF NOT EXISTS player_npc_email_questions_by_npc_date_idx
    ON player_npc_email_questions(server_id, target_npc_id, created_at, request_id);
CREATE INDEX IF NOT EXISTS player_npc_email_replies_by_npc_date_idx
    ON player_npc_email_questions(server_id, target_npc_id, ready_at, request_id)
    WHERE delivered_at IS NOT NULL AND reply_generated=true;

CREATE TABLE IF NOT EXISTS player_case_accusations (
    server_id uuid NOT NULL,
    account_id uuid NOT NULL REFERENCES player_accounts(id) ON DELETE CASCADE,
    case_id text NOT NULL,
    accused_npc_entity_id text NOT NULL DEFAULT '',
    crime_ids text[] NOT NULL DEFAULT '{}',
    board_id text NOT NULL,
    accepted boolean NOT NULL,
    outcome text NOT NULL CHECK (outcome IN ('arrested','insufficient')),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id,account_id,case_id,accused_npc_entity_id),
    FOREIGN KEY (server_id,case_id) REFERENCES cases(server_id,case_id) ON DELETE CASCADE
);
ALTER TABLE player_case_accusations ADD COLUMN IF NOT EXISTS accused_npc_entity_id text NOT NULL DEFAULT '';
ALTER TABLE player_case_accusations ADD COLUMN IF NOT EXISTS crime_ids text[] NOT NULL DEFAULT '{}';
ALTER TABLE player_case_accusations DROP CONSTRAINT IF EXISTS player_case_accusations_pkey;
ALTER TABLE player_case_accusations ADD PRIMARY KEY (server_id,account_id,case_id,accused_npc_entity_id);

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
    locked_at timestamptz,
    PRIMARY KEY (server_id, account_id, whiteboard_id)
);
ALTER TABLE player_whiteboards ADD COLUMN IF NOT EXISTS locked_at timestamptz;
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
DROP INDEX IF EXISTS player_whiteboard_one_accused_idx;
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

CREATE OR REPLACE FUNCTION reject_locked_whiteboard_changes() RETURNS trigger AS $$
DECLARE
    target_server uuid;
    target_account uuid;
    target_board text;
    board_locked boolean;
BEGIN
    IF TG_TABLE_NAME = 'player_whiteboards' THEN
        IF TG_OP = 'DELETE' AND OLD.locked_at IS NOT NULL THEN
            RAISE EXCEPTION 'submitted whiteboard is locked' USING ERRCODE = '55000';
        END IF;
        IF TG_OP = 'UPDATE' AND OLD.locked_at IS NOT NULL AND NEW.title IS DISTINCT FROM OLD.title THEN
            RAISE EXCEPTION 'submitted whiteboard is locked' USING ERRCODE = '55000';
        END IF;
        IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
        RETURN NEW;
    END IF;
    IF TG_OP = 'DELETE' THEN
        target_server := OLD.server_id;
        target_account := OLD.account_id;
        target_board := OLD.whiteboard_id;
    ELSE
        target_server := NEW.server_id;
        target_account := NEW.account_id;
        target_board := NEW.whiteboard_id;
    END IF;
    SELECT locked_at IS NOT NULL INTO board_locked FROM player_whiteboards
      WHERE server_id = target_server AND account_id = target_account AND whiteboard_id = target_board;
    IF board_locked THEN
        RAISE EXCEPTION 'submitted whiteboard is locked' USING ERRCODE = '55000';
    END IF;
    IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS player_whiteboards_lock_guard ON player_whiteboards;
CREATE TRIGGER player_whiteboards_lock_guard BEFORE UPDATE OR DELETE ON player_whiteboards
    FOR EACH ROW EXECUTE FUNCTION reject_locked_whiteboard_changes();
DROP TRIGGER IF EXISTS player_whiteboard_pins_lock_guard ON player_whiteboard_pins;
CREATE TRIGGER player_whiteboard_pins_lock_guard BEFORE INSERT OR UPDATE OR DELETE ON player_whiteboard_pins
    FOR EACH ROW EXECUTE FUNCTION reject_locked_whiteboard_changes();
DROP TRIGGER IF EXISTS player_whiteboard_connections_lock_guard ON player_whiteboard_connections;
CREATE TRIGGER player_whiteboard_connections_lock_guard BEFORE INSERT OR UPDATE OR DELETE ON player_whiteboard_connections
    FOR EACH ROW EXECUTE FUNCTION reject_locked_whiteboard_changes();

CREATE OR REPLACE FUNCTION reject_locked_whiteboard_note_changes() RETURNS trigger AS $$
BEGIN
    IF OLD.item_source = 'note' AND EXISTS (
        SELECT 1 FROM player_whiteboard_pins pin JOIN player_whiteboards board
          USING (server_id, account_id, whiteboard_id)
        WHERE pin.server_id = OLD.server_id AND pin.account_id = OLD.account_id
          AND pin.item_id = OLD.item_id AND board.locked_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'submitted whiteboard is locked' USING ERRCODE = '55000';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS player_whiteboard_notes_lock_guard ON player_discovered_items;
CREATE TRIGGER player_whiteboard_notes_lock_guard BEFORE UPDATE ON player_discovered_items
    FOR EACH ROW EXECUTE FUNCTION reject_locked_whiteboard_note_changes();
CREATE TABLE IF NOT EXISTS case_generation_schedule (
    server_id uuid PRIMARY KEY REFERENCES worlds(server_id) ON DELETE CASCADE,
    next_slot bigint NOT NULL DEFAULT 1,
    next_run_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Older case rows recover their company only through the case's explicit employment links.
UPDATE cases c SET company_id = (
    SELECT employment.employer_company_id
    FROM case_relationships relation
    JOIN employment_relationships employment
      ON employment.server_id = relation.server_id AND employment.relationship_id = relation.relationship_id
    WHERE relation.server_id = c.server_id AND relation.case_id = c.case_id
      AND relation.relationship_type = 'employment' AND relation.relationship_role = 'killer_employment'
    LIMIT 1
)
WHERE c.company_id IS NULL AND c.case_type = 'homicide'
  AND c.archetype_id IN ('workplace_conflict', 'financial_extortion');

INSERT INTO company_market_scores (server_id, company_id, score)
SELECT DISTINCT server_id, company_id, 1000 FROM cases
WHERE company_id IS NOT NULL AND archetype_id IN ('workplace_conflict', 'financial_extortion')
ON CONFLICT (server_id, company_id) DO NOTHING;

WITH impacts AS (
    SELECT server_id, case_id AS event_id, company_id,
           archetype_id AS event_type,
           CASE archetype_id WHEN 'financial_extortion' THEN -75 ELSE -25 END AS score_delta,
           COALESCE(crime_at, created_at) AS occurred_at,
           created_at AS recorded_at
    FROM cases
    WHERE company_id IS NOT NULL AND case_type IN ('homicide','financial_extortion')
      AND archetype_id IN ('workplace_conflict', 'financial_extortion')
), cumulative AS (
    SELECT impacts.*,
           GREATEST(500, 1000 + COALESCE(SUM(score_delta) OVER (
               PARTITION BY server_id, company_id ORDER BY recorded_at, event_id
               ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING
           ),0))::integer AS score_before,
           GREATEST(500, 1000 + SUM(score_delta) OVER (
               PARTITION BY server_id, company_id ORDER BY recorded_at, event_id
               ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
           ))::integer AS score_after
    FROM impacts
)
INSERT INTO company_market_events (server_id, event_id, company_id, event_type, score_delta, score_before, score_after, occurred_at, recorded_at)
SELECT server_id, event_id, company_id, event_type, score_delta,
       score_before, score_after, occurred_at, recorded_at
FROM cumulative
ON CONFLICT (server_id, event_id) DO UPDATE SET
    company_id = EXCLUDED.company_id, event_type = EXCLUDED.event_type,
    score_delta = EXCLUDED.score_delta, score_before = EXCLUDED.score_before,
    score_after = EXCLUDED.score_after, occurred_at = EXCLUDED.occurred_at,
    recorded_at = EXCLUDED.recorded_at;

UPDATE company_market_scores score SET score = latest.score_after, updated_at = latest.recorded_at
FROM (
    SELECT DISTINCT ON (server_id, company_id) server_id, company_id, score_after, recorded_at
    FROM company_market_events ORDER BY server_id, company_id, recorded_at DESC, event_id DESC
) latest
WHERE score.server_id = latest.server_id AND score.company_id = latest.company_id;
