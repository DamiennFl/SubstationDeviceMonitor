-- No specific schema, since data model is not large
-- Everything will go in public

-- +goose Up
CREATE TABLE devices (
    id                          UUID PRIMARY KEY, -- Devices unique by UUID
    name                        TEXT NOT NULL,
    ip                          TEXT NOT NULL,
    device_type                 TEXT NOT NULL,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ NULL,
    last_seen_at                TIMESTAMPTZ NULL,

    -- Define valid values for device_type
    CHECK (device_type IN ('relay'))
);

-- Partially unique index for device name. Only unique if not deleted
CREATE UNIQUE INDEX devices_name_active_key
ON devices (name)
WHERE deleted_at IS NULL;

-- Partially unique index for device ip. Only unique if not deleted
CREATE UNIQUE INDEX devices_ip_active_key
ON devices(ip)
WHERE deleted_at IS NULL;


CREATE TABLE rules (
    id              UUID PRIMARY KEY, -- Rule unique by ID
    device_id       UUID NOT NULL REFERENCES devices(id), -- Tied to device via foreign key
    name            TEXT NOT NULL,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    rule_type       TEXT NOT NULL,
    measurement     TEXT NOT NULL,
    comparison      TEXT,
    limit_value     DOUBLE PRECISION,
    lower_limit     DOUBLE PRECISION,
    upper_limit     DOUBLE PRECISION,
    deadband        DOUBLE PRECISION,
    trip_state      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (rule_type IN ('threshold', 'band', 'status')),
    CHECK (trip_state IN ('open', 'closed')),
    CHECK (comparison IN ('GT', 'LT')),

    CONSTRAINT check_valid_assigned_values CHECK ( -- Validate rules
        CASE
            WHEN rule_type = 'threshold' THEN comparison IS NOT NULL AND limit_value IS NOT NULL
            WHEN rule_type = 'band' THEN lower_limit IS NOT NULL AND upper_limit IS NOT NULL
            WHEN rule_type = 'status' THEN trip_state IS NOT NULL
            ELSE FALSE
        END
    ),

    UNIQUE (device_id, name) -- Rules by name are unique for a device, but two devices can have the same named rule
);

CREATE TABLE alarms (
    id              UUID PRIMARY KEY, -- Alarm unique by ID
    rule_id         UUID NULL REFERENCES rules(id) ON DELETE SET NULL, -- Tied to rule, until rule is deleted, then it just exists for history. Go validates the alarm is acknowledged.
    device_id       UUID NOT NULL REFERENCES devices(id), -- Tied to device
    raised_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    cleared_at      TIMESTAMPTZ NULL,
    clear_reason    TEXT,
    trigger_value   JSONB NOT NULL,
    rule_snapshot   JSONB NOT NULL,

    CHECK (clear_reason IN ('normal','ruleDisabled','ruleDeleted','deviceDeleted')),
    CHECK ((cleared_at IS NULL) = (clear_reason IS NULL))
);

CREATE INDEX device_alarms -- Non-unique index since one device can have multiple alarms. 
ON alarms (device_id); 

CREATE TABLE acknowledgements (
    alarm_id        UUID PRIMARY KEY REFERENCES alarms(id), -- One to one relationship between ack and alarm
    acknowledged_by TEXT NOT NULL,
    acknowledged_at TIMESTAMPTZ NOT NULL
);
 
CREATE TABLE readings (
    device_id       UUID NOT NULL REFERENCES devices(id), -- Readings map to device
    timestamp       TIMESTAMPTZ NOT NULL,
    current_a       DOUBLE PRECISION NOT NULL,
    current_b       DOUBLE PRECISION NOT NULL,
    current_c       DOUBLE PRECISION NOT NULL,
    voltage_a       DOUBLE PRECISION NOT NULL,
    voltage_b       DOUBLE PRECISION NOT NULL,
    voltage_c       DOUBLE PRECISION NOT NULL,
    frequency       DOUBLE PRECISION NOT NULL,
    real_power      DOUBLE PRECISION NOT NULL,
    reactive_power  DOUBLE PRECISION NOT NULL,
    breaker_status  TEXT NOT NULL,

    PRIMARY KEY (device_id, timestamp), -- Primary key is device_id and timestamp. Rejects dup readings.
    CHECK (breaker_status IN ('open', 'closed'))
);

-- +goose Down
-- Drop all tables
DROP TABLE readings;
DROP TABLE acknowledgements;
DROP TABLE alarms;
DROP TABLE rules;
DROP TABLE devices;