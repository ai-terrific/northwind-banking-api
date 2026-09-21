CREATE TABLE IF NOT EXISTS regulator_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id VARCHAR(255) NOT NULL UNIQUE,
    transfer_id UUID NOT NULL REFERENCES transfers(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deadline_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL '60 seconds'),
    delivered_at TIMESTAMP NULL,
    last_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_regulator_events_pending
    ON regulator_events(status, next_attempt_at);

CREATE INDEX idx_regulator_events_transfer_id
    ON regulator_events(transfer_id);

CREATE TABLE IF NOT EXISTS regulator_event_attempts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    regulator_event_id UUID NOT NULL REFERENCES regulator_events(id) ON DELETE CASCADE,
    attempt_number INTEGER NOT NULL,
    started_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP NOT NULL,
    response_status INTEGER NULL,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_regulator_event_attempts_event_id
    ON regulator_event_attempts(regulator_event_id, attempt_number);

CREATE TRIGGER update_regulator_events_updated_at BEFORE UPDATE ON regulator_events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
