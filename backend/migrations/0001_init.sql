-- No specific schema, since data model is not large
-- Everything will go in public

-- +goose Up
CREATE TABLE devices (
    id                          UUID PRIMARY KEY,
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
    id              UUID PRIMARY KEY,
    device_id       UUID NOT NULL REFERENCES devices(id),
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

    CONSTRAINT check_valid_assigned_values CHECK (
        CASE
            WHEN rule_type = 'threshold' THEN comparison IS NOT NULL AND limit_value IS NOT NULL
            WHEN rule_type = 'band' THEN lower_limit IS NOT NULL AND upper_limit IS NOT NULL
            WHEN rule_type = 'status' THEN trip_state IS NOT NULL
            ELSE FALSE
        END
    ),

    UNIQUE (device_id, name)
);

CREATE TABLE alarms (
    id              UUID PRIMARY KEY,
    rule_id         UUID NOT NULL REFERENCES rules(id),
    device_id       UUID NOT NULL REFERENCES devices(id),
    raised_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    cleared_at      TIMESTAMPTZ NULL,
    clear_reason    TEXT,
    trigger_value   TEXT NOT NULL, -- TODO: Validate this in Go since it can be number or text
    rule_snapshot   TEXT,

    CHECK (clear_reason IN ('normal','ruleDisabled','ruleDeleted','deviceDeleted')),
    CHECK ((cleared_at IS NULL) = (clear_reason IS NULL))
);

CREATE TABLE acknowledgements (
    alarm_id        UUID PRIMARY KEY REFERENCES alarms(id),
    acknowledged_by TEXT NOT NULL,
    acknowledged_at TIMESTAMPTZ NOT NULL
);
 
CREATE TABLE readings (
    device_id       UUID PRIMARY KEY REFERENCES devices(id),
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

    CHECK (breaker_status IN ('open', 'closed'))
);

-- +goose Down
DROP TABLE readings;
DROP TABLE acknowledgements;
DROP TABLE alarms;
DROP TABLE rules;
DROP TABLE devices;
