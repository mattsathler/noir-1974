CREATE TABLE IF NOT EXISTS game_servers (
    server_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS worlds (
    server_id uuid PRIMARY KEY REFERENCES game_servers(server_id) ON DELETE CASCADE,
    world_seed text NOT NULL CHECK (world_seed <> ''),
    generator_version integer NOT NULL CHECK (generator_version > 0),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS world_city_generators (
    server_id uuid NOT NULL REFERENCES worlds(server_id) ON DELETE CASCADE,
    city_id text NOT NULL CHECK (city_id <> ''),
    generator_version integer NOT NULL CHECK (generator_version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, city_id)
);

CREATE TABLE IF NOT EXISTS world_regions (
    server_id uuid NOT NULL REFERENCES worlds(server_id) ON DELETE CASCADE,
    region_id text NOT NULL CHECK (region_id <> ''),
    parent_region_id text,
    region_kind text NOT NULL CHECK (region_kind IN ('city', 'district')),
    display_name text NOT NULL CHECK (display_name <> ''),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, region_id),
    FOREIGN KEY (server_id, parent_region_id)
        REFERENCES world_regions(server_id, region_id)
        DEFERRABLE INITIALLY DEFERRED
);
CREATE INDEX IF NOT EXISTS world_regions_parent_idx
    ON world_regions(server_id, parent_region_id);

CREATE TABLE IF NOT EXISTS world_entities (
    server_id uuid NOT NULL,
    entity_id text NOT NULL CHECK (entity_id <> ''),
    entity_type text NOT NULL CHECK (entity_type IN ('npc', 'company', 'location')),
    region_id text NOT NULL,
    generation_key text NOT NULL CHECK (generation_key <> ''),
    location_kind text,
    materialized_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, entity_id),
    UNIQUE (server_id, region_id, generation_key),
    FOREIGN KEY (server_id, region_id)
        REFERENCES world_regions(server_id, region_id) ON DELETE CASCADE,
    CHECK (
        (entity_type = 'location' AND location_kind IS NOT NULL
            AND location_kind IN ('district', 'street', 'building', 'room'))
        OR (entity_type <> 'location' AND location_kind IS NULL)
    )
);
CREATE INDEX IF NOT EXISTS world_entities_region_idx
    ON world_entities(server_id, region_id, entity_type);

CREATE TABLE IF NOT EXISTS employment_relationships (
    server_id uuid NOT NULL,
    relationship_id text NOT NULL CHECK (relationship_id <> ''),
    employee_npc_id text NOT NULL,
    employer_company_id text NOT NULL,
    role text,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, relationship_id),
    UNIQUE (server_id, employee_npc_id, employer_company_id),
    FOREIGN KEY (server_id, employee_npc_id)
        REFERENCES world_entities(server_id, entity_id),
    FOREIGN KEY (server_id, employer_company_id)
        REFERENCES world_entities(server_id, entity_id),
    CHECK (employee_npc_id <> employer_company_id)
);
CREATE INDEX IF NOT EXISTS employment_by_employee_idx
    ON employment_relationships(server_id, employee_npc_id);
CREATE INDEX IF NOT EXISTS employment_by_company_idx
    ON employment_relationships(server_id, employer_company_id);

CREATE TABLE IF NOT EXISTS world_staff_email_directory (
    server_id uuid NOT NULL REFERENCES worlds(server_id) ON DELETE CASCADE,
    email text NOT NULL CHECK (email <> ''),
    npc_id text NOT NULL CHECK (npc_id <> ''),
    name text NOT NULL CHECK (name <> ''),
    role text NOT NULL DEFAULT '',
    status text NOT NULL DEFAULT 'alive',
    PRIMARY KEY (server_id, email),
    UNIQUE (server_id, npc_id, email)
);

CREATE TABLE IF NOT EXISTS personal_relationships (
    server_id uuid NOT NULL,
    relationship_id text NOT NULL CHECK (relationship_id <> ''),
    person_a_npc_id text NOT NULL,
    person_b_npc_id text NOT NULL,
    relationship_type text NOT NULL CHECK (
        relationship_type IN ('marriage', 'friendship', 'sibling', 'parent_child', 'rivalry')
    ),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, relationship_id),
    UNIQUE (server_id, person_a_npc_id, person_b_npc_id, relationship_type),
    FOREIGN KEY (server_id, person_a_npc_id)
        REFERENCES world_entities(server_id, entity_id),
    FOREIGN KEY (server_id, person_b_npc_id)
        REFERENCES world_entities(server_id, entity_id),
    CHECK (person_a_npc_id <> person_b_npc_id),
    CHECK (relationship_type = 'parent_child' OR person_a_npc_id < person_b_npc_id)
);
CREATE INDEX IF NOT EXISTS personal_relationships_by_a_idx
    ON personal_relationships(server_id, person_a_npc_id, relationship_type);
CREATE INDEX IF NOT EXISTS personal_relationships_by_b_idx
    ON personal_relationships(server_id, person_b_npc_id, relationship_type);

CREATE TABLE IF NOT EXISTS company_locations (
    server_id uuid NOT NULL,
    relationship_id text NOT NULL CHECK (relationship_id <> ''),
    company_id text NOT NULL,
    location_id text NOT NULL,
    location_role text NOT NULL CHECK (location_role IN ('headquarters', 'branch', 'office')),
    PRIMARY KEY (server_id, relationship_id),
    UNIQUE (server_id, company_id, location_id),
    FOREIGN KEY (server_id, company_id)
        REFERENCES world_entities(server_id, entity_id),
    FOREIGN KEY (server_id, location_id)
        REFERENCES world_entities(server_id, entity_id),
    CHECK (company_id <> location_id)
);
CREATE INDEX IF NOT EXISTS company_locations_by_location_idx
    ON company_locations(server_id, location_id);

CREATE TABLE IF NOT EXISTS world_mutations (
    mutation_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    mutation_order bigint GENERATED ALWAYS AS IDENTITY UNIQUE,
    server_id uuid NOT NULL,
    entity_id text NOT NULL,
    field_key text NOT NULL CHECK (field_key = 'status'),
    value jsonb NOT NULL CHECK (
        jsonb_typeof(value) = 'string' AND value #>> '{}' IN ('alive', 'missing', 'dead', 'arrested')
    ),
    created_at timestamptz NOT NULL DEFAULT now(),
    FOREIGN KEY (server_id, entity_id)
        REFERENCES world_entities(server_id, entity_id) ON DELETE CASCADE
);
-- Earlier schemas allowed only alive/dead. Keep existing QA and deployed tables in
-- sync with the status values already accepted by the world store.
ALTER TABLE world_mutations DROP CONSTRAINT IF EXISTS world_mutations_value_check;
ALTER TABLE world_mutations ADD CONSTRAINT world_mutations_value_check CHECK (
    jsonb_typeof(value) = 'string' AND value #>> '{}' IN ('alive', 'missing', 'dead', 'arrested')
);
CREATE INDEX IF NOT EXISTS world_mutations_by_entity_order_idx
    ON world_mutations(server_id, entity_id, mutation_order DESC);
